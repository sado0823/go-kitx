package main

import (
	"context"

	"github.com/sado0823/go-kitx/kit/log"
)

func init() {
	//kitx.New()
	//v := logrus.New()
	//v.Level = logrus.DebugLevel
	//logger := logrusV.New(v)
	//// fields & valuer
	//logger = log.WithFields(logger,
	//	"service.name", "hellworld",
	//	"service.version", "v1.0.0",
	//	"ts", log.DefaultTimestamp,
	//	"caller", log.DefaultCaller,
	//)
	//
	//production, _ := zap.NewProduction(zap.AddCallerSkip(3))
	//logger = zapV.New(production)
	//
	//logger = log.WithFields(logger,
	//	"service.name", "hellworld",
	//	"service.version", "v1.0.0",
	//	"ts", log.DefaultTimestamp,
	//	"caller", log.DefaultCaller,
	//)
	//log.SetGlobal(logger)
}

func main() {

	log.Debug("debug", 123)
	log.Info("info", 456)
	log.Warn("warn")
	log.Error("error")
	log.Fatal("fatal")
	log.Context(context.Background()).Error("ccccccc")
}
