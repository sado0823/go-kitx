# go-kitx



Some practical tools、 algorithms written in Go


- [x] [**p2c grpc balancer**](https://github.com/sado0823/go-kitx/tree/master/grpc/balancer/p2c)
```go
// example

```

- [x] [**ast rule engine**](https://github.com/sado0823/go-kitx/blob/master/kit/rule/parser.go)
```go
// example
import (
    . "github.com/sado0823/go-kitx/kit/rule"
)

func main(){
	// use build in custom func `in`
    expr := `func in(foo,"a",1,1.2)`
    param := map[string]interface{}{
        "foo": "a",
        "in":  1,
    }
    
    res, err := Do(context.Background(), expr, param)
    if err != nil {
        panic(err)
    }
}

```