package pbhelloworld

//go:generate protoc -I . -I ../../../third_party --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. --go-http-kitx_out=paths=source_relative:. --validate_out=paths=source_relative,lang=go:. ./helloworld.proto


// go:generate protoc -I . -I ../../../third_party --go_out=paths=source_relative:. --go-errors-kitx_out=paths=source_relative:. ./helloworld.errors.proto

// go install github.com/sado0823/go-kitx/cmd/protoc-gen-go-http-kitx@latest
// go install github.com/sado0823/go-kitx/cmd/protoc-gen-go-errors-kitx@latest
// go install github.com/envoyproxy/protoc-gen-validate@latest