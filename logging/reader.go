package logging

import (
	"bufio"
	"bytes"
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

	err := r.searchFile(r.filesInfo[readFrom])
	if err != nil {
		return err
	}

	// read one file in reverse order and parse the log lines to check for datetime
	// last written log in the file equals to the file ModTime()
	//err := r.streamOne(r.filesInfo[readFrom])
	//if err != nil {
	//	return err
	//}

	//others := r.filesInfo[readFrom+1 : len(r.filesInfo)]
	//for _, fi := range others {
	//	chunks := r.stream(fi)
	//	for c := range chunks {
	//		if c.err != nil {
	//			return c.err
	//		}
	//
	//		_, err := fmt.Fprintln(w, c.line)
	//		if err != nil {
	//			return err
	//		}
	//	}
	//}

	return nil
}

func (r *LogReader) searchFile(fi fileInfo) error {
	filePath := path.Join(r.cfg.Directory, fi.name)
	file, err := os.Open(filePath)
	defer func() {
		_ = file.Close()
	}()
	if err != nil {
		return err
	}

	pos := int64(0)
	scanLines := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		advance, token, err = bufio.ScanLines(data, atEOF)
		pos += int64(advance)
		return
	}

	start := int64(io.SeekStart)
	end := fi.size
	mid := start + (end-start)/2
	var prevTime time.Time
	var prevOffset int64
	for {
		//file.Seek(start, io.SeekStart)
		//SeekLine(file, 0, io.SeekCurrent)
		scanner := bufio.NewScanner(file)
		scanner.Split(scanLines)
		scanner.Scan()
		line := scanner.Text()
		fmt.Println(line)
		t, err := r.parseLogDateTime(line)
		if err != nil {
			return err
		}

		nowMinusT := time.Now().UTC().Add(time.Duration(-r.cfg.LastNMinutes) * time.Minute)
		if nowMinusT.Sub(t) <= 0 {
			end = mid
			mid = start + (end-start)/2
			fmt.Println("up", mid)
		} else {
			start = mid
			mid = start + (end-start)/2
			fmt.Println("down", mid)
		}

		_, err = file.Seek(mid, io.SeekStart)
		if err != nil {
			return err
		}
		offSet, err := SeekLine(file, 0, io.SeekCurrent)
		if err != nil {
			return err
		}

		prevOffset = offSet
		if t.Sub(prevTime) >= 0 {
			break
		}

		if start == mid || end == mid {
			break
		}

		prevTime = t
	}

	_, err = file.Seek(prevOffset, io.SeekStart)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(scanLines)
	scanner.Scan()
	fmt.Println(scanner.Text())

	return nil
}

func (r *LogReader) streamOne(fi fileInfo) error {
	filePath := path.Join(r.cfg.Directory, fi.name)
	file, err := os.Open(filePath)
	defer func() {
		_ = file.Close()
	}()
	if err != nil {
		return err
	}

	//offset, err := file.Seek(0, io.SeekStart)
	//if err != nil {
	//	return err
	//}
	var offset int64
	buffSize := int64(100)
	buff := &bytes.Buffer{}
	for {
		bs := make([]byte, buffSize)
		_, err = file.ReadAt(bs, offset)
		if errors.Is(err, io.EOF) {
			buff.Write(bs)
			line, _ := buff.ReadString('\n')
			fmt.Printf(line)
			break
		}
		if err != nil {
			return err
		}

		offset += buffSize
		buff.Write(bs)

		l, _ := buff.ReadString('\n')
		t, err := r.parseLogDateTime(l)
		if err != nil {
			return err
		}
		fmt.Println(t)
	}

	//fmt.Print(buff.String())

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
