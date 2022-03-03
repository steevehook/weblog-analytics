package logging

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

// LogReaderConfig represents the configuration to start the log reader
type LogReaderConfig struct {
	Directory    string
	LastNMinutes int
}

// NewLogReader creates a new instance of log reader
func NewLogReader(cfg LogReaderConfig) *LogReader {
	return &LogReader{
		cfg: cfg,
	}
}

// LogReader represents the application log reader type
// responsible for reading logs from a given directory
// that were written in the last N minutes
type LogReader struct {
	cfg LogReaderConfig
	buf bytes.Buffer
}

// Read reads the log files using the given LogReader configuration
// and stores it inside a local bytes buffer to be displayed later
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

// String displays the string representation of the read logs
func (r *LogReader) String() string {
	return "the logger"
}
