# Weblog Analytics

### Prerequisites

- Make sure you have installed Go version >= `1.17`

### Build

```shell
# compiles and generates binaries for log-reader and log-generator inside the ./bin directory
make build
``` 

### Run

```shell
# run: "make build" first
# only run once to generate the test data, it may take a while
./bin/log-generator
# run the log-reader with the specified cli arguments
./bin/log-reader -d <path/to/log/files> -t <last_n_minutes>
# run the program directory without generating any binary
go run cmd/log-generator/main.go
go run cmd/log-reader/main.go -d <path/to/log/files> -t <last_n_minutes>
# display all logs from testdata directory that happened in the last 5 minutes
./bin/log-reader -d ./testdata -t 5
```

### Test

```shell
# runs all the tests present in test files
make test
# runs all the benchmarks present in test files
make bench
```
