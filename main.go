package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path"
	"regexp"
	"time"
)

func main() {
	filePath := path.Join("testdata", "logs.txt")
	f, err := os.Open(filePath)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		log.Fatal(err)
	}

	file := newLogFile(f)
	offset, err := file.search(2)
	if err != nil {
		log.Fatal(err)
	}

	if offset == -1 {
		return
	}
	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

const dateTimeGroupName = "datetime"

var errInvalidLogFormat = errors.New("invalid log format")

func newLogFile(file *os.File) *logFile {
	logFormat := fmt.Sprintf(
		`^(\S+) (\S+) (\S+) \[(?P<%s>[\w:/]+\s[+\-]\d{4})\] "(\S+)\s?(\S+)?\s?(\S+)?" (\d{3}|-) (\d+|-)\s?"?([^"]*)"?\s?"?([^"]*)?"?$`,
		dateTimeGroupName,
	)
	return &logFile{
		File:  file,
		regEx: regexp.MustCompile(logFormat),
	}
}

type logFile struct {
	*os.File
	regEx *regexp.Regexp
}

func (file *logFile) search(lastNMinutes uint) (int64, error) {
	var top, bottom, pos, prevPos, offset, prevOffset int64
	scanLines := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		advance, token, err = bufio.ScanLines(data, atEOF)
		prevPos = pos
		pos += int64(advance)
		return
	}

	stat, err := os.Stat(file.Name())
	if err != nil {
		return -1, err
	}
	bottom = stat.Size()
	nowMinusT := time.Now().UTC().Add(-time.Duration(lastNMinutes) * time.Minute)
	var prevLogTime time.Time
	for top <= bottom {
		// define the middle relative to top and bottom positions
		middle := top + (bottom-top)/2
		//fmt.Println("new top", top, "new bottom", bottom)
		//fmt.Println("middle", middle)
		// seek the file at the middle
		_, err := file.Seek(middle, io.SeekStart)
		if err != nil {
			return -1, err
		}
		// reposition the middle to the beginning of the line
		offset, err = file.seekLine(0, io.SeekCurrent)
		if err != nil {
			return -1, err
		}

		// scan 1 line and convert it to int
		scanner := bufio.NewScanner(file)
		scanner.Split(scanLines)
		scanner.Scan()
		line := scanner.Text()
		if line == "" {
			// we'll consider an empty line an EOF
			break
		}

		logTime, err := file.parseLogDateTime(line)
		if err != nil {
			return -1, err
		}

		if nowMinusT.Sub(logTime) > 0 {
			// the starting log is way down (relative to the middle)
			// move down the top
			top = offset + (pos - prevPos)
		} else if prevLogTime.Sub(logTime) < 0 {
			// the starting log is way up (relative to the middle)
			// move up the bottom
			bottom = offset - (pos - prevPos)
		} else if nowMinusT.Sub(prevLogTime) < 0 && offset != top {
			return top, nil
		}

		if offset == top {
			return offset - (pos - prevPos), nil
		}
		if offset == bottom {
			return bottom, nil
		}
		prevLogTime = logTime
		prevOffset = offset
	}

	if nowMinusT.Minute() == prevLogTime.Minute() {
		return prevOffset, nil
	}

	return -1, nil
}

func (file *logFile) seekLine(lines int64, whence int) (int64, error) {
	const bufferSize = 32 * 1024 // 32KB
	buf := make([]byte, bufferSize)
	bufLen := 0
	lines = int64(math.Abs(float64(lines)))
	seekBack := lines < 1
	lineCount := int64(0)

	// seekBack ignores the first match lines == 0
	// then goes to the beginning of the current line
	if seekBack {
		lineCount = -1
	}

	pos, err := file.Seek(0, whence)
	left := pos
	offset := int64(bufferSize * -1)
	for b := 1; ; b++ {
		if seekBack {
			// on seekBack 2nd buffer onward needs to seek
			// past what was just read plus another buffer size
			if b == 2 {
				offset *= 2
			}

			// if next seekBack will pass beginning of file
			// buffer is 0 to unread position
			if pos+offset <= 0 {
				buf = make([]byte, left)
				left = 0
				pos, err = file.Seek(0, io.SeekStart)
			} else {
				left = left - bufferSize
				pos, err = file.Seek(offset, io.SeekCurrent)
			}
		}
		if err != nil {
			break
		}

		bufLen, err = file.Read(buf)
		if err != nil {
			return file.Seek(0, io.SeekEnd)
		}
		for i := 0; i < bufLen; i++ {
			idx := i
			if seekBack {
				idx = bufLen - i - 1
			}
			if buf[idx] == '\n' {
				lineCount++
			}
			if lineCount == lines {
				if seekBack {
					return file.Seek(int64(i)*-1, io.SeekCurrent)
				}
				return file.Seek(int64(bufLen*-1+i+1), io.SeekCurrent)
			}
		}
		if seekBack && left == 0 {
			return file.Seek(0, io.SeekStart)
		}
	}

	return pos, err
}

// apache common log example
// 127.0.0.1 user-identifier frank [06/Mar/2022:05:30:00 +0000] "GET /api/endpoint HTTP/1.0" 500 123
func (file *logFile) parseLogDateTime(l string) (time.Time, error) {
	matches := file.regEx.FindStringSubmatch(l)
	if len(matches) == 0 {
		return time.Time{}, errInvalidLogFormat
	}

	var dateTime string
	for i, name := range file.regEx.SubexpNames() {
		if name == dateTimeGroupName {
			dateTime = matches[i]
			break
		}
	}
	if dateTime == "" {
		return time.Time{}, errInvalidLogFormat
	}

	dateTimeFormat := "02/Jan/2006:15:04:05 -0700"
	t, err := time.Parse(dateTimeFormat, dateTime)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}
