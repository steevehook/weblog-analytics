package logging

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type logReaderSuite struct {
	suite.Suite
}

func (s *logReaderSuite) SetupSuite() {
}

func (s *logReaderSuite) SetupTest() {
}

func (s *logReaderSuite) Test_NewLogReader() {
}

func TestLogReader(t *testing.T) {
	suite.Run(t, new(logReaderSuite))
}

func BenchmarkLogReader(b *testing.B) {
	b.ReportAllocs()
}
