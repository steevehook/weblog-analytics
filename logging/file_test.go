package logging

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type fileSuite struct {
	suite.Suite
	dataDir string
}

func (s *fileSuite) SetupSuite() {
	s.dataDir = "testdata"
	s.Require().NoError(os.RemoveAll(s.dataDir))
	s.Require().NoError(os.Mkdir(s.dataDir, 0777))
}

func (s *fileSuite) TearDownSuite() {
	s.Require().NoError(os.RemoveAll(s.dataDir))
}

func (s *fileSuite) Test_NewFile() {
	f, err := os.CreateTemp(s.dataDir, "*")
	defer func() { s.Require().NoError(f.Close()) }()
	s.Require().NoError(err)

	file := NewFile(f)

	s.NotNil(file)
}

func (s *fileSuite) Test_LogFile_seekLine() {
	data := "some\ntest\nstring\n"
	f, err := os.CreateTemp(s.dataDir, "*")
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
	log := `127.0.0.1 user-identifier frank [06/Mar/2022:05:30:00 +0000] "GET /api/endpoint HTTP/1.0" 500 123`
	expectedTime, err := time.Parse(dateTimeFormat, "06/Mar/2022:05:30:00 +0000")
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
	tests := []struct{
		name string
		log string
		expectedErr string
	}{
		{
			name: "Empty LogLine",
			log: "",
			expectedErr: "invalid log format",
		},
		{
			name: "Invalid LogLine",
			log: "this log line is not valid",
			expectedErr: "invalid log format",
		},
		{
			name: "Invalid DateFormat",
			log: `127.0.0.1 user-identifier frank [36/Mar/2022:05:30:00 +0000] "GET /api/endpoint HTTP/1.0" 500 123`,
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

func BenchmarkSearch(b *testing.B) {
	b.ReportAllocs()
}
