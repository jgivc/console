package transport

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TimeoutReaderSuite struct {
	suite.Suite
	reader *MockReader
}

func (suite *TimeoutReaderSuite) SetupTest() {
	suite.reader = new(MockReader)
}

func (suite *TimeoutReaderSuite) TestRead_Successful() {
	expectedData := []byte("Hello, TimeoutReader!")
	suite.reader.On("Read", mock.Anything).Return(len(expectedData), nil).Run(func(args mock.Arguments) {
		b, ok := args.Get(0).([]byte)
		if !ok {
			suite.T().Fatal("cannot convert")
		}
		copy(b, expectedData)
	}).Once()
	suite.reader.On("Read", mock.Anything).Return(0, io.EOF).After(100 * time.Millisecond)

	timeout := 1 * time.Second
	bufferSize := len(expectedData)
	tr := newTimeoutReader(context.Background(), suite.reader, timeout, bufferSize)

	data := make([]byte, bufferSize)
	n, err := tr.Read(data)

	suite.NoError(err, "Unexpected error")
	suite.Equal(expectedData, data[:n], "Data mismatch")

	suite.reader.AssertExpectations(suite.T())
}

func (suite *TimeoutReaderSuite) TestRead_Timeout() {
	suite.reader.On("Read", mock.Anything).Return(0, nil)

	timeout := 1 * time.Millisecond
	bufferSize := 1024
	tr := newTimeoutReader(context.Background(), suite.reader, timeout, bufferSize)

	data := make([]byte, bufferSize)
	_, err := tr.Read(data)
	suite.ErrorIs(err, os.ErrDeadlineExceeded)
	suite.reader.AssertExpectations(suite.T())
}

func (suite *TimeoutReaderSuite) TestRead_Error() {
	expectedErr := errors.New("mock read error")
	suite.reader.On("Read", mock.Anything).Return(0, expectedErr)

	timeout := 1 * time.Second
	bufferSize := 1024
	tr := newTimeoutReader(context.Background(), suite.reader, timeout, bufferSize)

	data := make([]byte, bufferSize)
	_, err := tr.Read(data)

	suite.Require().Error(err)
	suite.EqualError(err, "read error: mock read error", "Error mismatch")
	suite.reader.AssertExpectations(suite.T())
}

func (suite *TimeoutReaderSuite) TestSetTimeout() {
	timeout := 2 * time.Second
	bufferSize := 1024
	tr := newTimeoutReader(context.Background(), suite.reader, timeout, bufferSize)

	tm := 200 * time.Microsecond

	tr.SetTimeout(tm)
	suite.Equal(tm, tr.(*timeoutReaderImpl).timeout)
}

func (suite *TimeoutReaderSuite) TestClose() {
	timeout := 2 * time.Second
	bufferSize := 1024
	tr := newTimeoutReader(context.Background(), suite.reader, timeout, bufferSize)

	err := tr.Close()

	suite.NoError(err, "Unexpected error")

	data := make([]byte, 1)
	_, err2 := tr.Read(data)
	suite.Error(err2, "Expected read error after close")
}

func TestTimeoutReaderSuite(t *testing.T) {
	suite.Run(t, new(TimeoutReaderSuite))
}

type MockReader struct {
	mock.Mock
}

func (m *MockReader) Read(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}
