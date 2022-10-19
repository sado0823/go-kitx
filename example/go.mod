module github.com/sado0823/go-kitx/example

go 1.16

require (
	github.com/sado0823/go-kitx v0.0.2
	github.com/sado0823/go-kitx/plugin/registry/etcd v0.0.0-20221009073440-50faf22007b7
	go.etcd.io/etcd/client/v3 v3.5.5
)

require go.uber.org/zap v1.23.0 // indirect

replace (
	github.com/sado0823/go-kitx => ../
	github.com/sado0823/go-kitx/plugin/logger/logrus => ../plugin/logger/logrus
	github.com/sado0823/go-kitx/plugin/logger/zap => ../plugin/logger/zap
)
