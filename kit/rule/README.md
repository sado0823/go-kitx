# ast rule engine


##### ExampleDo
```go
import (
    "context"
    "fmt"
    "unicode/utf8"
    
    "github.com/sado0823/go-kitx/kit/rule"
)

func ExampleDo() {

    params := map[string]interface{}{"foo": 1}
    
    value, err := rule.Do(context.Background(), `foo + 1`, params)
    if err != nil {
        fmt.Println(err)
    }
    
    fmt.Print(value)
    
    // Output:
    // 2
}
```


##### ExampleNew
```go
import (
    "context"
    "fmt"
    "unicode/utf8"
    
    "github.com/sado0823/go-kitx/kit/rule"
)

func ExampleNew() {
    params := map[string]interface{}{"foo": 1}
    
    parser, err := rule.New(context.Background(), `foo + 1`)
    if err != nil {
        fmt.Println(err)
    }
    
    value, err := parser.Eval(params)
    if err != nil {
        fmt.Println(err)
    }
    
    fmt.Print(value)
    
    // Output:
    // 2
}
```

##### ExampleWithCustomFn
```go
import (
    "context"
    "fmt"
    "unicode/utf8"
    
    "github.com/sado0823/go-kitx/kit/rule"
)

func ExampleWithCustomFn() {
    params := map[string]interface{}{"foo": 1}
    
    value, err := rule.Do(
		
        context.Background(),
		
        `func in(foo,2,"abc",1) && func strlen("abc") == 3 && func isTrue(true) && func isTrue(false) == false`,
        
        params,
		
        /* custom func `strlen` return args[0]'s count with float64 type */
        rule.WithCustomFn("strlen", func(arguments ...interface{}) (interface{}, error) {
            if len(arguments) == 0 {
                return 0, nil
            }
            return float64(utf8.RuneCount([]byte(arguments[0].(string)))), nil
        }),
		
         /*custom func `isTrue` return if args[0] is true with bool type*/
        rule.WithCustomFn("isTrue", func(arguments ...interface{}) (interface{}, error) {
            if len(arguments) == 0 {
                return 0, nil
            }
            return arguments[0].(bool) == true, nil
        }),
    )
    if err != nil {
        fmt.Println(err)
    }
    
    fmt.Print(value)
    
    // Output:
    // true
}
```