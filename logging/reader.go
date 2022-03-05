package logging

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"time"
)

const (
	dateTimeGroupName = "datetime"
	dateTimeFormat    = "02/Jan/2006:15:04:05 -0700"
)

var (
	logFormat = fmt.Sprintf(
		`^(\S+) (\S+) (\S+) \[(?P<%s>[\w:/]+\s[+\-]\d{4})\] "(\S+)\s?(\S+)?\s?(\S+)?" (\d{3}|-) (\d+|-)\s?"?([^"]*)"?\s?"?([^"]*)?"?$`,
		dateTimeGroupName,
	)
	errInvalidLogFormat = errors.New("invalid log format")
)

type fileInfo struct {
	name    string
	modTime time.Time
	size    int64
}

// LogReaderConfig represents the configuration to start the log reader
type LogReaderConfig struct {
	Directory    string
	LastNMinutes int
}

// NewLogReader creates a new instance of log reader
func NewLogReader(cfg LogReaderConfig) (*LogReader, error) {
	files, err := ioutil.ReadDir(cfg.Directory)
	if err != nil {
		return nil, err
	}

	filesInfo := make([]fileInfo, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fi := fileInfo{
			name:    file.Name(),
			modTime: file.ModTime().UTC(),
			size:    file.Size(),
		}
		filesInfo = append(filesInfo, fi)
	}

	lr := &LogReader{
		cfg:       cfg,
		regEx:     regexp.MustCompile(logFormat),
		filesInfo: filesInfo,
	}
	return lr, nil
}

// LogReader represents the application log reader type
// responsible for reading logs from a given directory
// that were written in the last N minutes
type LogReader struct {
	cfg       LogReaderConfig
	filesInfo []fileInfo
	regEx     *regexp.Regexp
}

// Read reads the log files using the given LogReader configuration
// and stores it inside a local bytes buffer to be displayed later
func (r *LogReader) Write(ctx context.Context, w io.Writer) error {
	select {
	case <-ctx.Done():
		return nil
	default:
		return r.write(w)
	}
}

// if there are an infinite number of log files,
// knowing the exact log rotation period may help
// skip iterations up to the very close of the log file
func (r *LogReader) write(w io.Writer) error {
	// if the r.cfg.LastNMinutes < log rotation period
	// start reading from the last file
	readFrom := -1
	for i, fi := range r.filesInfo {
		// if current time in UTC minus LastNMinutes => we may have multiple log files to read
		nowMinusT := time.Now().UTC().Add(time.Duration(-r.cfg.LastNMinutes) * time.Minute)
		if nowMinusT.Sub(fi.modTime) <= 0 {
			readFrom = i
			break
		}
	}

	if readFrom == -1 {
		return nil
	}

	// search call

	// read one file in reverse order and parse the log lines to check for datetime
	// last written log in the file equals to the file ModTime()
	//err := r.streamOne(r.filesInfo[readFrom])
	//if err != nil {
	//	return err
	//}

	others := r.filesInfo[readFrom+1 : len(r.filesInfo)]
	for _, fi := range others {
		chunks := r.stream(fi)
		for c := range chunks {
			if c.err != nil {
				return c.err
			}

			_, err := fmt.Fprintln(w, c.line)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type chunk struct {
	line string
	err  error
}

func (r *LogReader) stream(fi fileInfo) chan chunk {
	out := make(chan chunk)
	go func() {
		filePath := path.Join(r.cfg.Directory, fi.name)
		file, err := os.Open(filePath)
		defer func() {
			_ = file.Close()
		}()
		if err != nil {
			out <- chunk{
				err:  err,
				line: "",
			}
			return
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			out <- chunk{
				err:  nil,
				line: scanner.Text(),
			}
		}

		close(out)
	}()
	return out
}

func (r *LogReader) parseLogDateTime(l string) (time.Time, error) {
	matches := r.regEx.FindStringSubmatch(l)
	if len(matches) == 0 {
		return time.Time{}, errInvalidLogFormat
	}

	var dateTime string
	for i, name := range r.regEx.SubexpNames() {
		if name == dateTimeGroupName {
			dateTime = matches[i]
			break
		}
	}
	if dateTime == "" {
		return time.Time{}, errInvalidLogFormat
	}

	t, err := time.Parse(dateTimeFormat, dateTime)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}
