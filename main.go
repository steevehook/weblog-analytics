package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"

	"github.com/steevehook/weblog-analytics/logging"
)

func main() {
	lastNMinutes := 6
	nowMinusT := time.Now().UTC().Add(-time.Duration(lastNMinutes) * time.Minute)

	filePath := path.Join("testdata", "http-10.log")
	f, err := os.Open(filePath)
	defer func() {_ = f.Close()}()
	if err != nil {
		log.Fatal(err)
	}

	file := logging.NewFile(f)
	fmt.Println(nowMinusT)
	offset, err := file.IndexTime(nowMinusT)
	if err != nil {
		log.Fatal(err)
	}

	_, err = f.Seek(offset, io.SeekStart)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}
