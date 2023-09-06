package transport

import (
	"context"
	"io"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	testDataDir            = "testdata"
	testDummyFileName      = "dummy_config.xml"
	testDummyFileNotExists = "notextsts"
)

type DummyTransportTestSuite struct {
	suite.Suite
	fileName string
	t        *dummyTransport
}

func (suite *DummyTransportTestSuite) SetupTest() {
	suite.fileName = path.Join(testDataDir, testDummyFileName)
	suite.t = new(dummyTransport)
	suite.t.fileName = suite.fileName
}

func (suite *DummyTransportTestSuite) TestSuccess() {
	suite.t.timeout = 500 * time.Millisecond
	err := suite.t.Open(context.Background(), nil)
	suite.NoError(err)
	b := make([]byte, 1024)

	for _, sd := range suite.t.dr.rd.SendData {
		n, err2 := suite.t.Read(b)
		suite.ErrorIs(err2, io.EOF)
		suite.Equal(len(sd.Send), n)
		suite.Equal(sd.Send, b[:n])
	}
}

func (suite *DummyTransportTestSuite) TestNotSetTimeout() {
	err := suite.t.Open(context.Background(), nil)
	suite.Error(err)
}

func (suite *DummyTransportTestSuite) TestOpenFail() {
	suite.fileName = testDummyFileNotExists
	err := suite.t.Open(context.Background(), nil)
	suite.Error(err)
}

func (suite *DummyTransportTestSuite) TestTimeout() {
	suite.t.timeout = 100 * time.Millisecond
	err := suite.t.Open(context.Background(), nil)
	suite.NoError(err)
	b := make([]byte, 1024)

	_, err = suite.t.Read(b)
	suite.ErrorIs(err, os.ErrDeadlineExceeded)
}

func TestDummyTransportTestSuite(t *testing.T) {
	suite.Run(t, new(DummyTransportTestSuite))
}
