package logging

import (
	"io"
	"os"
	"testing"

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
			name:           "Line Zero",
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

func TestLogFile(t *testing.T) {
	suite.Run(t, new(fileSuite))
}

func BenchmarkSearch(b *testing.B) {
	b.ReportAllocs()
}
