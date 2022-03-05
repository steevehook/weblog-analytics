package logging

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
)

func search(f *os.File, search int64) (int64, error) {
	var top, bottom, pos, prevPos int64
	scanLines := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		advance, token, err = bufio.ScanLines(data, atEOF)
		prevPos = pos
		pos += int64(advance)
		return
	}

	file := newLineFile(f)
	stat, err := os.Stat(file.Name())
	if err != nil {
		return -1, err
	}
	bottom = stat.Size()

	for top <= bottom {
		// define the middle relative to top and bottom positions
		middle := top + (bottom-top)/2
		// seek the file at the middle
		_, err := file.Seek(middle, io.SeekStart)
		if err != nil {
			return -1, err
		}
		// reposition the middle to the beginning of the line
		offset, err := file.seekLine(0, io.SeekCurrent)
		if err != nil {
			return -1, err
		}

		// scan 1 line and convert it to int
		scanner := bufio.NewScanner(file)
		scanner.Split(scanLines)
		scanner.Scan()
		line := scanner.Text()
		if line == "" {
			// we'll consider this an EOF
			// so let's break
			break
		}

		num, err := strconv.Atoi(line)
		if err != nil {
			// we only want to look up sorted numbers
			// if there's anything that's not valid just error
			return -1, fmt.Errorf("invalid line at offset %d : %w", offset, err)
		}

		// found the number, return the offset
		if int64(num) == search {
			return offset, nil
		}
		if int64(num) > search {
			// the number is way up (relative to the middle)
			// move up the bottom
			bottom = offset - (pos - prevPos)
		} else {
			// the number is way down (relative to the middle)
			// move down the top
			top = offset + (pos - prevPos)
		}
	}

	// the number was not found
	return -1, nil
}


type lineFile struct {
	*os.File
}

func newLineFile(file *os.File) *lineFile {
	return &lineFile{file}
}

func (file *lineFile) seekLine(lines int64, whence int) (int64, error) {
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
