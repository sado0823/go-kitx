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
