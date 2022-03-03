# Weblog Analytics

### Prerequisites

- Make sure you have installed Go version >= `1.17`

### Build

```shell
# compiles and generates a binary for log-reader inside the ./bin directory
make build
``` 

### Run

```shell
# run: "make build" first
./bin/log-reader -d <path/to/log/files> -t <last_n_minutes>
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
