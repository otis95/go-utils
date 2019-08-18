### Utils help lib

Use reflect pkg to do something.

### How to use

- TransformStruct

```
type User struct {
    Username string
    Age      int
}
type UserViewModel struct {
    Username string
}
var (
    user = User{
        "lumore",
        25,
    }
    userView UserViewModel
)

utils.TransformStruct(&user, &userView)

fmt.Println(userView)
```

- CompareSlice

```
type User struct {
    Username string
    Age      int
}
var (
    aUserList = []User{
        {
            "lumore",
            25,
        },
        {
            "a",
            25,
        },
    }
    bUserList = []User{
        {
            "a",
            25,
        },
    }
)
addSlice, removeSlice, _ := utils.CompareSlice(aUserList, bUserList)
fmt.Println(addSlice)
fmt.Println(removeSlice)
```