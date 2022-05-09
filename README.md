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