module github.com/sado0823/go-kitx/example

go 1.17

require github.com/sirupsen/logrus v1.8.1

require (
	github.com/sado0823/go-kitx v0.0.2-0.20220918181943-5591528ae40a
	github.com/sado0823/go-kitx/plugin/logger/logrus v0.0.1
	github.com/sado0823/go-kitx/plugin/logger/zap v0.0.1
	go.uber.org/zap v1.23.0
)

require (
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/sys v0.0.0-20220808155132-1c4a2a72c664 // indirect
)

replace (
	github.com/sado0823/go-kitx => ../
	github.com/sado0823/go-kitx/plugin/logger/logrus => ../plugin/logger/logrus
	github.com/sado0823/go-kitx/plugin/logger/zap => ../plugin/logger/zap
)
