# Weblog Analytics

### Prerequisites

- Make sure you have installed Go version >= `1.17`

### Build

```shell
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
make test
```
