package logging

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

type LogReaderConfig struct {
	Directory    string
	LastNMinutes int
}

func NewLogReader(cfg LogReaderConfig) *LogReader {
	return &LogReader{
		cfg: cfg,
	}
}

type LogReader struct {
	cfg LogReaderConfig
	buf bytes.Buffer
}

func (r *LogReader) Read(ctx context.Context) *LogReader {
	select {
	case <-ctx.Done():
		return r
	default:
		time.Sleep(time.Second * 5)
		fmt.Println("processing is done")
	}
	return r
}

func (r *LogReader) String() string {
	return "the logger"
}
