package reflect

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
)

var (
	tagParser = make(map[string]func(source interface{}) interface{})

	ParamKindErr     = errors.New("param kind need to be same")
	SliceElemKindErr = errors.New("slice elem kind need to be struct")
	SliceKindErr     = errors.New("param need to be slice kind")
)

func RegisterTagParser(name string, parser func(source interface{}) interface{}) {
	if _, exist := tagParser[name]; exist {
		panic(fmt.Sprintf("%s parser is exsit.", name))
	}
	tagParser[name] = parser
}

// TransformStruct : A struct => B struct
// The same paramName will be transform.
func TransformStruct(src, to interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
		}
	}()

	t := reflect.TypeOf(src).Elem()
	v := reflect.ValueOf(src).Elem()

	// find reflect value by param name.
	finderValue := func(paramName string, where interface{}) (reflect.Value, bool) {
		value := reflect.ValueOf(where).Elem().FieldByName(paramName)

		if reflect.DeepEqual(value, reflect.Value{}) {
			return value, false
		}

		return value, true
	}

	// select all params with anonymous parms
	params := []string{}

	for i := 0; i < v.NumField(); i++ {
		switch v.Field(i).Type().Kind() {
		case reflect.Struct:
			if t.Field(i).Anonymous {
				for j := 0; j < t.Field(i).Type.NumField(); j++ {
					params = append(params, t.FieldByIndex([]int{i, j}).Name)
				}
				continue
			}
			fallthrough
		default:
			params = append(params, t.Field(i).Name)
		}
	}

	for _, param := range params {
		var (
			paramValue   = v.FieldByName(param)
			paramType, _ = t.FieldByName(param)
			toValue      reflect.Value
			exist        bool
		)

		// find param from to struct
		// if not find param, continue
		toValue, exist = finderValue(param, to)
		if !exist {
			continue
		}
		if toValue.Type().Kind() != paramValue.Type().Kind() {
			return ParamKindErr
		}

		switch paramValue.Type().Kind() {
		case reflect.Struct:
			// recursive
			if err := TransformStruct(paramValue.Addr().Interface(), toValue.Addr().Interface()); err != nil {
				return err
			}
		case reflect.Slice:
			switch paramType.Type.Elem().Kind() {
			case reflect.String:
				if parser, ok := tagParser[paramType.Tag.Get("tag")]; !ok {
					goto SetValue
				} else {
					toValue.Set(reflect.ValueOf(parser(paramValue.Interface())))
				}
			case reflect.Struct:
				typ := toValue.Type().Elem()
				if typ.Kind() != reflect.Struct {
					return SliceElemKindErr
				}
				MigrateSlice(paramValue.Addr().Interface(), toValue.Addr().Interface())
			default:
				goto SetValue
			}
		default:
			goto SetValue
		}
	SetValue:
		if toValue.IsValid() && paramValue.Type() == toValue.Type() {
			toValue.Set(paramValue)
		}
	}

	return nil
}

// CompareSlice : compare A slice with B silice.
// Return: 1. Additional elements slice 2.Removed elements slice 3. error
func CompareSlice(newSlice, oldSlice interface{}) (addSlice, removeSlice []interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
		}
	}()

	var (
		// nsv => new slice value
		nsv = reflect.ValueOf(newSlice)
		// osv => old slice value
		osv = reflect.ValueOf(oldSlice)
	)

	switch {
	case nsv.Kind() != reflect.Slice:
		return nil, nil, SliceKindErr
	case osv.Kind() != reflect.Slice:
		return nil, nil, SliceKindErr
	}

	diff := func(nsv, osv reflect.Value) []interface{} {
		var diffSlice []interface{}
		for i := 0; i < nsv.Len(); i++ {
			var isfind bool
			for j := 0; j < osv.Len(); j++ {
				if reflect.DeepEqual(nsv.Index(i).Interface(), osv.Index(j).Interface()) {
					isfind = true
					break
				}
			}
			if !isfind {
				diffSlice = append(diffSlice, nsv.Index(i).Interface())
			}
		}
		return diffSlice
	}

	return diff(nsv, osv), diff(osv, nsv), nil
}

// MigrateSlice : A slice => B slice
// The same paramName will be transform.
func MigrateSlice(src interface{}, to interface{}) {
	switch {
	case reflect.TypeOf(src).Elem().Kind() != reflect.Slice:
		return
	case reflect.TypeOf(to).Elem().Kind() != reflect.Slice:
		return
	}

	srcValue := reflect.ValueOf(src).Elem()
	toValue := reflect.ValueOf(to).Elem()
	typ := toValue.Type().Elem()

	for j := 0; j < srcValue.Len(); j++ {
		obj := reflect.New(typ)
		TransformStruct(srcValue.Index(j).Addr().Interface(), obj.Interface())
		toValue.Set(reflect.Append(toValue, obj.Elem()))
	}
}

// SumSliceParamsValue : Count slice's param value
// Input: 1. slice 2. struct paramName 3. count result ptr
func SumSliceParamsValue(src interface{}, paramName string, sum *int32) {
	if reflect.TypeOf(src).Kind() != reflect.Slice {
		return
	}
	srcValue := reflect.ValueOf(src)
	for j := 0; j < srcValue.Len(); j++ {
		var (
			v          = srcValue.Index(j)
			paramValue = v.FieldByName(paramName)
		)
		switch paramValue.Kind() {
		case reflect.Int:
			*sum += int32(paramValue.Interface().(int))
		case reflect.Int32:
			*sum += paramValue.Interface().(int32)
		case reflect.String:
			count, err := strconv.Atoi(paramValue.Interface().(string))
			if err != nil {
				continue
			}
			*sum += int32(count)
		}
	}
}

// FilterSlice : filter silice that you not want
// Input: 1. slice 2. paramName 3. param value
func FilterSlice(src interface{}, paramName string, want interface{}, filterOrNot ...bool) {
	switch {
	case reflect.TypeOf(src).Elem().Kind() != reflect.Slice:
		return
	case reflect.TypeOf(src).Kind() != reflect.Ptr:
		return
	}

	filter := true
	if len(filterOrNot) != 0 {
		filter = filterOrNot[0]
	}

	srcValue := reflect.ValueOf(src).Elem()
	for i := 0; i < srcValue.Len(); {
		var (
			v          = srcValue.Index(i)
			paramValue = v.FieldByName(paramName)
		)
		if reflect.DeepEqual(paramValue.Interface(), want) == filter {
			srcValue.Set(reflect.AppendSlice(srcValue.Slice(0, i), srcValue.Slice(i+1, srcValue.Len())))
		} else {
			i++
		}
	}
}
