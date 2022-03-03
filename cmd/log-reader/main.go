package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/steevehook/weblog-analytics/logging"
)

func main() {
	quit := make(chan os.Signal, 1)
	directoryFlag := flag.String("d", ".", "the directory where all the logs are stored")
	minutesFlag := flag.Int("t", 1, "last n minutes of worth of logs to read")

	flag.Parse()
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	cfg := logging.LogReaderConfig{
		Directory:    *directoryFlag,
		LastNMinutes: *minutesFlag,
	}
	logReader := logging.NewLogReader(cfg)

	go func() {
		fmt.Println(logReader.Read(ctx).String())
		quit <- os.Interrupt
	}()

	select {
	case <-quit:
		cancel()
	}
}
