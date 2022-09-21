package main


import (
	"context"

	"github.com/sado0823/go-kitx/kit/log"
	pLogger "github.com/sado0823/go-kitx/plugin/logger/logrus"

	"github.com/sirupsen/logrus"
)

func init() {
	v := logrus.New()
	v.Level = logrus.DebugLevel
	logger := pLogger.New(v)
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
	log.Fatal("fatal")
	log.Context(context.Background()).Error("ccccccc")
}