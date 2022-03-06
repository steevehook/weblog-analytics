package main

import "fmt"

func main() {
	// example of apache log
	// 127.0.0.1 user-identifier frank [04/Mar/2022:00:03:00 +0000] "GET /api/endpoint HTTP/1.0" 500 123
	fmt.Println("generate couple of log files")
	fmt.Println("create a giant 1GB log file")
	fmt.Println("add couple of thousands inside every other file")
}
