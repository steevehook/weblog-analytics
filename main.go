package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path"
	"strconv"
)

func main() {
	filePath := path.Join("testdata", "numbers.txt")
	file, err := os.Open(filePath)
	defer func() {
		_ = file.Close()
	}()
	if err != nil {
		log.Fatal(err)
	}

	offset, err := binarySearch(file, 14)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(offset)
}

func binarySearch(f *os.File, search int64) (int64, error) {
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
		offset, err := file.SeekLine(0, io.SeekCurrent)
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

const BufferLength = 32 * 1024

type lineFile struct {
	*os.File
}

func newLineFile(file *os.File) *lineFile {
	return &lineFile{file}
}

func (file *lineFile) SeekLine(lines int64, whence int) (int64, error) {
	position, err := file.Seek(0, whence)
	buf := make([]byte, BufferLength)
	bufLen := 0
	lineSep := byte('\n')
	seekBack := lines < 1
	lines = int64(math.Abs(float64(lines)))
	matchCount := int64(0)

	// seekBack ignores first match
	// allows 0 to go to begining of current line
	if seekBack {
		matchCount = -1
	}

	leftPosition := position
	offset := int64(BufferLength * -1)

	for b := 1; ; b++ {
		if err != nil {
			break
		}

		if seekBack {

			// on seekBack 2nd buffer onward needs to seek
			// past what was just read plus another buffer size
			if b == 2 {
				offset *= 2
			}

			// if next seekBack will pass beginning of file
			// buffer is 0 to unread position
			if position+int64(offset) <= 0 {
				buf = make([]byte, leftPosition)
				position, err = file.Seek(0, io.SeekStart)
				leftPosition = 0
			} else {
				position, err = file.Seek(offset, io.SeekCurrent)
				leftPosition = leftPosition - BufferLength
			}
		}
		if err != nil {
			break
		}

		bufLen, err = file.Read(buf)
		if err != nil {
			break
		} else if seekBack && leftPosition == 0 {
			err = io.EOF
		}

		for i := 0; i < bufLen; i++ {
			iToCheck := i
			if seekBack {
				iToCheck = bufLen - i - 1
			}
			byteToCheck := buf[iToCheck]

			if byteToCheck == lineSep {
				matchCount++
			}

			if matchCount == lines {
				if seekBack {
					return file.Seek(int64(i)*-1, io.SeekCurrent)
				}
				return file.Seek(int64(bufLen*-1+i+1), io.SeekCurrent)
			}
		}
	}

	if err == io.EOF && !seekBack {
		position, _ = file.Seek(0, io.SeekEnd)
	} else if err == io.EOF && seekBack {
		position, _ = file.Seek(0, io.SeekStart)

		// no io.EOF err on SeekLine(0,0)
		if lines == 0 {
			err = nil
		}
	}

	return position, err
}
