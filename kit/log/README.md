# Logger

## Usage

### 全局共用一个global logger

```go
package main

import (
	"github.com/sado0823/go-kitx/kit/log"
)

func main() {
	log.Debug("debug")
	log.Info("info")
	log.Warn("warn")
	log.Error("error")
	log.Fatal("fatal")

	// output
	//DEBUG  ts=2022-09-12T19:11:58+08:00 caller=go-kitx/main.go:8 msg=debug
	//INFO  ts=2022-09-12T19:11:58+08:00 caller=go-kitx/main.go:9 msg=info
	//WARN  ts=2022-09-12T19:11:58+08:00 caller=go-kitx/main.go:10 msg=warn
	//ERROR  ts=2022-09-12T19:11:58+08:00 caller=go-kitx/main.go:11 msg=error
	//FATAL  ts=2022-09-12T19:11:58+08:00 caller=go-kitx/main.go:12 msg=fatal

}

```

### 每个package自定义自己前缀的logger

```go
package main

import (
	"fmt"
	"os"

	"github.com/sado0823/go-kitx/kit/log"
)

var (
	logger log.Logger
	helper *log.Helper
)

func init() {

	logger = log.NewStd(os.Stdout)
	// fields & valuer
	logger = log.WithFields(logger,
		"service.name", "hellworld",
		"service.version", "v1.0.0",
		"ts", log.DefaultTimestamp,
		"caller", log.Caller(3),
	)

	// helper
	helper = log.NewHelper(logger)
}

func main() {
	logger.Log(log.LevelError, "error")
	logger.Log(log.LevelInfo, "info")

	fmt.Println("")

	helper.Debug("debug")
	helper.Info("info")
	helper.Warn("warn")
	helper.Error("error")
	helper.Fatal("fatal")

	// output
	//ERROR  service.name=hellworld service.version=v1.0.0 ts=2022-09-12T19:09:59+08:00 caller=go-kitx/main.go:30 error=log key value unpaired
	//INFO  service.name=hellworld service.version=v1.0.0 ts=2022-09-12T19:09:59+08:00 caller=go-kitx/main.go:31 info=log key value unpaired
	//
	//DEBUG  service.name=hellworld service.version=v1.0.0 ts=2022-09-12T19:09:59+08:00 caller=log/helper.go:47 msg=debug
	//INFO  service.name=hellworld service.version=v1.0.0 ts=2022-09-12T19:09:59+08:00 caller=log/helper.go:59 msg=info
	//WARN  service.name=hellworld service.version=v1.0.0 ts=2022-09-12T19:09:59+08:00 caller=log/helper.go:71 msg=warn
	//ERROR  service.name=hellworld service.version=v1.0.0 ts=2022-09-12T19:09:59+08:00 caller=log/helper.go:83 msg=error
	//FATAL  service.name=hellworld service.version=v1.0.0 ts=2022-09-12T19:09:59+08:00 caller=log/helper.go:95 msg=fatal
}

```

### 其它方式

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sado0823/go-kitx/kit/log"
)

var (
	logger  = log.NewHelper(log.WithFields(log.NewStd(os.Stdout), "caller", log.Caller(4)))
	logger2 = log.NewHelper(log.WithFields(log.GetGlobal(), "dao.name", "article"))
)

func main() {
	logger.Log(log.LevelError, "abc1", "123")
	logger2.Log(log.LevelError, "abc1", "123")
	log.Log(log.LevelError, "abc2", "432")

	fmt.Println("")

	logger.Log(log.LevelError, "abc1", "123")
	logger2.Log(log.LevelError, "abc1", "123")
	log.Log(log.LevelError, "abc2", "432")

	fmt.Println("")

	logger.Error("1")
	logger2.Error("1")
	log.Error("2")

	fmt.Println("")
	ctx := context.Background()

	logger.WithContext(ctx).Error("1")
	logger2.WithContext(ctx).Error("1")
	log.Context(ctx).Error("222")

	// output
	//ERROR  caller=go-kitx/main.go:17 abc1=123
	//ERROR  ts=2022-09-12T19:06:09+08:00 caller=go-kitx/main.go:18 dao.name=article abc1=123
	//ERROR  ts=2022-09-12T19:06:09+08:00 caller=go-kitx/main.go:19 abc2=432
	//
	//ERROR  caller=go-kitx/main.go:23 abc1=123
	//ERROR  ts=2022-09-12T19:06:09+08:00 caller=go-kitx/main.go:24 dao.name=article abc1=123
	//ERROR  ts=2022-09-12T19:06:09+08:00 caller=go-kitx/main.go:25 abc2=432
	//
	//ERROR  caller=go-kitx/main.go:29 msg=1
	//ERROR  ts=2022-09-12T19:06:09+08:00 caller=go-kitx/main.go:30 dao.name=article msg=1
	//ERROR  ts=2022-09-12T19:06:09+08:00 caller=go-kitx/main.go:31 msg=2
	//
	//ERROR  caller=go-kitx/main.go:36 msg=1
	//ERROR  ts=2022-09-12T19:06:09+08:00 caller=go-kitx/main.go:37 dao.name=article msg=1
	//ERROR  ts=2022-09-12T19:06:09+08:00 caller=go-kitx/main.go:38 msg=222

}

```


## Third party log plugin

### logrus

```shell
go get -u github.com/sado0823/go-kitx/plugin/logger/logrus
```

```go
import (
	"context"
	"fmt"
	
	"github.com/sado0823/go-kitx/kit/log"
	pLogger "github.com/sado0823/go-kitx/plugin/logger/logrus"
	"github.com/sirupsen/logrus"
)

func init() {
	logger := pLogger.New(logrus.New())
	// fields & valuer
	logger = log.WithFields(logger,
		"service.name", "hellworld",
		"service.version", "v1.0.0",
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
	)
	
	log.SetGlobal(logger)
}

func main() {

	log.Debug("debug", 123)
	log.Info("info", 456)
	log.Warn("warn")
	log.Error("error")
	//log.Fatal("fatal")
	log.Context(context.Background()).Error("ccccccc")
}
```