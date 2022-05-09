# ast rule engine

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