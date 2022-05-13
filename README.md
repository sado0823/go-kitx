# go-kitx



Some practical tools、 algorithms written in Go


- [x] [**p2c grpc balancer**](https://github.com/sado0823/go-kitx/tree/master/grpc/balancer/p2c)
```go
// example
func test() {
    cc, err := grpc.Dial(r.Scheme()+":///test.server",
        grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingConfig": [{"%s":{}}]}`, p2c.Name)))
    if err != nil {
        t.Fatalf("failed to dial: %v", err)
    }
    defer cc.Close()
}
```

- [x] [**ast rule engine**](https://github.com/sado0823/go-kitx/tree/master/kit/rule)

__supported operator__ 

* **comparator**: `>` `>=` `<` `<=` `==`

* **bitwise**: `&` `|` `^`

* **bitwiseShift**: `<<` `>>`

* **additive**: `+` `-`

* **multiplicative**: `*` `/` `%`

* **prefix**: `!`(NOT)  `-`(NEGATE)

* **logic**: `&&` `||`

* **others**: `(` `)` `,` `func`(do func call with build in function and custom function)

* **params type**: `Ident` `Number` `String` `Bool` (DO Not support `array` `func` `struct`)


```go
// example
import (
    . "github.com/sado0823/go-kitx/kit/rule"
)

func main(){

    expr := `foo + 1 > bar`
	
    param := map[string]interface{}{
        "foo": 5,
        "bar": 6,
    }
	
    res, err := Do(context.Background(), expr, param)
    if err != nil {
        panic(err)
    }
}

```