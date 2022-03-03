build:
	@echo "building the log-reader binary"
	go build -o bin/log-reader cmd/log-reader/main.go

test:
	@echo "running all tests"
	go test -v ./...
