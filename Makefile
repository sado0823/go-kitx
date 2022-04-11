.PHONY: test fmt


test:
	go test ./...

fmt:
	go fmt ./...
	go vet ./...


