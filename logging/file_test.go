package logging

import (
	"fmt"
	"io"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const dataDir = "testdata"

type fileSuite struct {
	suite.Suite
}

func (s *fileSuite) SetupSuite() {
	s.Require().NoError(os.RemoveAll(dataDir))
	s.Require().NoError(os.Mkdir(dataDir, 0777))
}

func (s *fileSuite) TearDownSuite() {
	s.Require().NoError(os.RemoveAll(dataDir))
}

func (s *fileSuite) Test_NewFile() {
	f, err := os.CreateTemp(dataDir, "*")
	defer func() { s.Require().NoError(f.Close()) }()
	s.Require().NoError(err)

	file := NewFile(f)

	s.NotNil(file)
}

func (s *fileSuite) Test_LogFile_IndexTime_Success() {
	// add some old logs as well
	now := time.Now().UTC()
	// add some old logs
	logs := "127.0.0.1 user-identifier frank [02/Mar/2022:05:30:00 +0000] \"GET /api/endpoint HTTP/1.0\" 500 123\n"
	logs += "127.0.0.1 user-identifier frank [02/Mar/2022:05:35:00 +0000] \"GET /api/endpoint HTTP/1.0\" 500 123\n"
	// add some new logs
	numberOfNewLogs := 7
	for i := 0; i < numberOfNewLogs; i++ {
		logs += fmt.Sprintf(
			"127.0.0.1 user-identifier frank [%v] \"GET /api/endpoint HTTP/1.0\" 500 123\n",
			now.Add(-time.Duration(numberOfNewLogs-i)*time.Minute).Format(dateTimeFormat),
		)
	}
	f, err := os.CreateTemp(dataDir, "*-http.log")
	defer func() { s.Require().NoError(f.Close()) }()
	s.Require().NoError(err)
	_, err = f.WriteString(logs)
	s.Require().NoError(err)
	file := NewFile(f)
	s.NotNil(file)

	lineLen := int64(97)
	for i := 0; i < numberOfNewLogs; i++ {
		lookupTime := time.Now().UTC().Add(-time.Duration(numberOfNewLogs-i) * time.Minute)
		offset, err := file.IndexTime(lookupTime)
		expectedOffset := lineLen*int64(i) + int64(i)
		oldOffset := lineLen*2 + 2

		s.NoError(err)
		s.Equal(expectedOffset, offset-oldOffset)
	}
}

func (s *fileSuite) Test_LogFile_IndexTime_TinyDifference() {
	// prove the very small difference in seconds
}

func (s *fileSuite) Test_LogFile_IndexTime_NoLogs() {
	now := time.Now().UTC()
	numberOfLogs := 7
	logs := ""
	for i := 0; i < numberOfLogs; i++ {
		logs += fmt.Sprintf(
			"127.0.0.1 user-identifier frank [%v] \"GET /api/endpoint HTTP/1.0\" 500 123\n",
			now.Add(-time.Duration(numberOfLogs-i)*time.Hour).Format(dateTimeFormat),
		)
	}
	f, err := os.CreateTemp(dataDir, "*-http.log")
	defer func() { s.Require().NoError(f.Close()) }()
	s.Require().NoError(err)
	_, err = f.WriteString(logs)
	s.Require().NoError(err)
	file := NewFile(f)
	s.NotNil(file)

	lookupTime := time.Now().UTC().Add(-1 * time.Minute)
	offset, err := file.IndexTime(lookupTime)

	s.NoError(err)
	s.Equal(int64(-1), offset)
}

func (s *fileSuite) Test_LogFile_IndexTime_Error() {
	logs := "some invalid log line\n"
	f, err := os.CreateTemp(dataDir, "*-http.log")
	defer func() { s.Require().NoError(f.Close()) }()
	s.Require().NoError(err)
	_, err = f.WriteString(logs)
	s.Require().NoError(err)
	file := NewFile(f)
	s.NotNil(file)

	lookupTime := time.Now().UTC().Add(-1 * time.Minute)
	offset, err := file.IndexTime(lookupTime)

	s.EqualError(err, "invalid log format")
	s.Equal(int64(-1), offset)
}

func (s *fileSuite) Test_LogFile_seekLine() {
	data := "some\ntest\nstring\n"
	f, err := os.CreateTemp(dataDir, "*")
	defer func() { s.Require().NoError(f.Close()) }()
	s.Require().NoError(err)
	_, err = f.WriteString(data)
	s.Require().NoError(err)

	_, err = f.Seek(8, io.SeekStart)
	s.NoError(err)
	file := NewFile(f)
	s.NotNil(file)

	tests := []struct {
		name           string
		lines          int64
		whence         int
		expectedOffset int64
	}{
		{
			name:           "Line Zero CurrentLine",
			lines:          0,
			whence:         io.SeekCurrent,
			expectedOffset: 5,
		},
		{
			name:           "LinesGreaterThanZero SeekCurrent",
			lines:          3,
			whence:         io.SeekCurrent,
			expectedOffset: int64(len(data)),
		},
		{
			name:           "LinesGreaterThanZero SeekEnd",
			lines:          3,
			whence:         io.SeekEnd,
			expectedOffset: int64(len(data)),
		},
		{
			name:           "LinesGreaterThanZero SeekStart LastLine",
			lines:          3,
			whence:         io.SeekStart,
			expectedOffset: int64(len(data)),
		},
		{
			name:           "LineGreaterThanZero SeekStart SecondLine",
			lines:          2,
			whence:         io.SeekStart,
			expectedOffset: 10,
		},
		{
			name:           "NegativeLines SeekStart SecondLine",
			lines:          -2,
			whence:         io.SeekStart,
			expectedOffset: 10,
		},
	}
	for _, test := range tests {
		s.Run(test.name, func() {
			offset, err := file.seekLine(test.lines, test.whence)

			s.NoError(err)
			s.Equal(test.expectedOffset, offset)
		})
	}
}

func (s *fileSuite) Test_LogFile_parseLogTime_Success() {
	log := `127.0.0.1 user-identifier frank [04/Mar/2022:05:30:00 +0000] "GET /api/endpoint HTTP/1.0" 500 123`
	expectedTime, err := time.Parse(dateTimeFormat, "04/Mar/2022:05:30:00 +0000")
	s.Require().NoError(err)
	file := NewFile(nil)
	s.NotNil(file)

	t, err := file.parseLogTime(log)

	s.NoError(err)
	s.True(t.Equal(expectedTime))
}

func (s *fileSuite) Test_LogFile_parseLogTime_Error() {
	file := NewFile(nil)
	s.NotNil(file)
	tests := []struct {
		name        string
		log         string
		expectedErr string
	}{
		{
			name:        "Empty LogLine",
			log:         "",
			expectedErr: "invalid log format",
		},
		{
			name:        "Invalid LogLine",
			log:         "this log line is not valid",
			expectedErr: "invalid log format",
		},
		{
			name:        "Invalid DateFormat",
			log:         `127.0.0.1 user-identifier frank [36/Mar/2022:05:30:00 +0000] "GET /api/endpoint HTTP/1.0" 500 123`,
			expectedErr: `parsing time "36/Mar/2022:05:30:00 +0000": day out of range`,
		},
	}
	for _, test := range tests {
		s.Run(test.name, func() {
			t, err := file.parseLogTime(test.log)

			s.EqualError(err, test.expectedErr)
			s.True(t.IsZero())
		})
	}
}

func TestLogFile(t *testing.T) {
	suite.Run(t, new(fileSuite))
}

// generate the big log file to be able to benchmark properly
func BenchmarkSearch(b *testing.B) {
	f, err := os.Open(path.Join(dataDir, "http-1.log"))
	defer func() { assert.NoError(b, f.Close()) }()
	assert.NoError(b, err)
	file := NewFile(f)
	assert.NotNil(b, file)
	b.ResetTimer()

	// look up last minutes in ascending
	// we don't care about the offset, we only want to benchmark
	// and check for execution time and memory footprint
	lookupTime := time.Now().UTC().Add(-27 * time.Minute)
	_, err = file.IndexTime(lookupTime)

	assert.NoError(b, err)
	b.ReportAllocs()
}
