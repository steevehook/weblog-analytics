package logging

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type readerSuite struct {
	suite.Suite
}

func (s *readerSuite) SetupSuite() {
}

func (s *readerSuite) SetupTest() {
}

func (s *readerSuite) Test_NewLogReader() {
}

func TestLogReader(t *testing.T) {
	suite.Run(t, new(readerSuite))
}

func BenchmarkLogReader(b *testing.B) {
	b.ReportAllocs()
}
