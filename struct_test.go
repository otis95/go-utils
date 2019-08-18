package reflect

import (
	"reflect"
	"testing"
)

func TestDiffSlice(t *testing.T) {
	type testModel struct {
		name string
		age  int
	}
	var testCase = []struct {
		id          int
		newSlice    interface{}
		oldSlice    interface{}
		addSlice    []interface{}
		deleteSlice []interface{}
	}{
		{
			1,
			[]string{"222", "333", "4444"},
			[]string{"111", "555", "4444"},
			[]interface{}{"222", "333"},
			[]interface{}{"111", "555"},
		},
		{
			2,
			[]string{"222", "333", "4444"},
			[]string{},
			[]interface{}{"222", "333", "4444"},
			[]interface{}{},
		},
		{
			3,
			[]string{},
			[]string{"222", "333", "4444"},
			[]interface{}{},
			[]interface{}{"222", "333", "4444"},
		},
		{
			4,
			[]int{3, 4, 5},
			[]int{3, 4, 6},
			[]interface{}{5},
			[]interface{}{6},
		},
		{
			5,
			[]testModel{
				testModel{
					name: "test1",
					age:  21,
				},
				testModel{
					name: "test2",
					age:  22,
				},
			},
			[]testModel{
				testModel{
					name: "test3",
					age:  23,
				},
				testModel{
					name: "test2",
					age:  22,
				},
			},
			[]interface{}{
				testModel{
					name: "test1",
					age:  21,
				},
			},
			[]interface{}{
				testModel{
					name: "test3",
					age:  23,
				},
			},
		},
		{
			6,
			[]int{3, 4, 5, 6},
			[]int{3, 4, 5},
			[]interface{}{6},
			[]interface{}{},
		},
	}

	for _, test := range testCase {
		addSlice, deleteSlice, _ := CompareSlice(test.newSlice, test.oldSlice)
		switch {
		case len(addSlice) == 0 && len(test.addSlice) == 0:
			break
		case !reflect.DeepEqual(addSlice, test.addSlice):
			t.Errorf("%d test failed. Add slice Want => %v , Get => %v", test.id, test.addSlice, addSlice)
		}
		switch {
		case len(deleteSlice) == 0 && len(test.deleteSlice) == 0:
			break
		case !reflect.DeepEqual(deleteSlice, test.deleteSlice):
			t.Errorf("%d test failed. Delete slice Want => %v , Get => %v", test.id, test.deleteSlice, deleteSlice)
		}
	}
}
func TestTransformStruct(t *testing.T) {
	type c struct {
		Name string
	}
	type cwant struct {
		Name string
	}
	type testModel struct {
		Name    string
		ID      int
		Next    c
		IntList []int
		C       []c
	}
	type testWantModel struct {
		Name    string
		ID      int
		Next    cwant
		C       []cwant
		IntList []int
	}

	var testCase = []struct {
		id    int
		input testModel
		want  testWantModel
	}{
		{
			1, testModel{
				Name: "test",
				ID:   1,
				Next: c{
					Name: "test",
				},
				C: []c{
					c{
						Name: "test",
					},
					c{
						Name: "test",
					},
				},
				IntList: []int{111, 222},
			},
			testWantModel{
				Name: "test",
				ID:   1,
				Next: cwant{
					Name: "test",
				},
				C: []cwant{
					cwant{
						Name: "test",
					},
					cwant{
						Name: "test",
					},
				},
				IntList: []int{111, 222},
			},
		},
	}
	for _, test := range testCase {
		var to testWantModel
		TransformStruct(&test.input, &to)
		if !reflect.DeepEqual(to, test.want) {
			t.Errorf("test failed. Want => %v , Get => %v", test.want, to)
		}
	}
}

func TestSumSliceParamsValue(t *testing.T) {
	type testModel struct {
		A string
		B int
		C string
	}
	var testCase = []struct {
		id    int
		input interface{}
		want  int32
	}{
		{
			1,
			[]testModel{
				{"1", 1, "1"},
				{"2", 2, "2"},
			},
			3,
		},
	}
	for _, test := range testCase {
		var sum int32
		SumSliceParamsValue(test.input, "B", &sum)
		if sum != test.want {
			t.Errorf("test failed. Want => %v , Get => %v", test.want, sum)
		}
	}
}

func TestFilterSlice(t *testing.T) {
	type testModel struct {
		A string
		B int
		C string
	}
	var testCase = []struct {
		id          int
		input       []testModel
		filterName  string
		filterValue interface{}
		filter      bool
		want        []testModel
	}{
		{
			1,
			[]testModel{
				{"1", 1, "1"},
				{"2", 2, "2"},
				{"3", 2, "3"},
			},
			"B",
			2,
			true,
			[]testModel{
				{"1", 1, "1"},
			},
		},
		{
			1,
			[]testModel{
				{"1", 1, "1"},
				{"2", 2, "2"},
				{"3", 2, "3"},
			},
			"B",
			2,
			false,
			[]testModel{
				{"2", 2, "2"},
				{"3", 2, "3"},
			},
		},
	}
	for _, test := range testCase {
		var sum int32
		if test.filter {
			FilterSlice(&test.input, test.filterName, test.filterValue)
		} else {
			FilterSlice(&test.input, test.filterName, test.filterValue, false)
		}
		if !reflect.DeepEqual(test.input, test.want) {
			t.Errorf("test failed. Want => %v , Get => %v", test.want, sum)
		}
	}
}
