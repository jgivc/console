package util

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TimeoutReaderMock struct {
	mock.Mock
}

func (m *TimeoutReaderMock) Read(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *TimeoutReaderMock) SetReadTimeout(t time.Duration) {
	m.Called(t)
}

func (m *TimeoutReaderMock) Close() error {
	return nil
}

type PromptReaderSuite struct {
	suite.Suite
	timeoutReaderMock *TimeoutReaderMock
	matchLength       int
	BuffZize          int
}

func (suite *PromptReaderSuite) SetupTest() {
	suite.matchLength = 10
	suite.BuffZize = 512
	suite.timeoutReaderMock = new(TimeoutReaderMock)
}

func (suite *PromptReaderSuite) TestReadWithPrompt() {
	expectedData := []byte(`Authorization required!

	Username: `)
	suite.timeoutReaderMock.On("Read", mock.Anything).Return(len(expectedData), nil).Run(func(args mock.Arguments) {
		b, ok := args.Get(0).([]byte)
		if !ok {
			suite.T().Fatal("cannot convert")
		}
		copy(b, expectedData)
	}).Once()
	// suite.timeoutReaderMock.On("Read", mock.Anything).Return(0, io.EOF).Once()

	pr := NewPromptReader(suite.timeoutReaderMock, suite.BuffZize, suite.matchLength)
	err := pr.SetPromptPattern(`(?i)user\w+:\s+`)
	suite.NoError(err)
	buffer := make([]byte, len(expectedData))
	n, err := pr.Read(buffer)
	suite.NoError(err)
	suite.Equal(n, len(expectedData))
	suite.Equal(buffer, expectedData)
	suite.timeoutReaderMock.AssertExpectations(suite.T())
}

func (suite *PromptReaderSuite) TestReadWithPromptDeadLine() {
	suite.timeoutReaderMock.On("Read", mock.Anything).Return(0, os.ErrDeadlineExceeded).Once()
	expectedData := []byte(`Authorization required!

	Username: `)
	suite.timeoutReaderMock.On("Read", mock.Anything).Return(len(expectedData), nil).Run(func(args mock.Arguments) {
		b, ok := args.Get(0).([]byte)
		if !ok {
			suite.T().Fatal("cannot convert")
		}
		copy(b, expectedData)
	}).Once()
	// suite.timeoutReaderMock.On("Read", mock.Anything).Return(0, io.EOF).Once()

	pr := NewPromptReader(suite.timeoutReaderMock, suite.BuffZize, suite.matchLength)
	err := pr.SetPromptPattern(`(?i)user\w+:\s+`)
	suite.NoError(err)
	pr.SetDeadLine(time.Now().Add(2 * time.Second))
	buffer := make([]byte, len(expectedData))
	n, err := pr.Read(buffer)
	suite.NoError(err)
	suite.Equal(n, len(expectedData))
	suite.Equal(buffer, expectedData)
	suite.timeoutReaderMock.AssertExpectations(suite.T())
}

func (suite *PromptReaderSuite) TestReadWithPromptPartsDeadLine() {
	suite.timeoutReaderMock.On("Read", mock.Anything).Return(0, os.ErrDeadlineExceeded).Once()
	expectedData := []byte(`Authorization required!

	Username: `)
	part1 := expectedData[:len(expectedData)/2]
	suite.timeoutReaderMock.On("Read", mock.Anything).Return(len(part1), nil).Run(func(args mock.Arguments) {
		b, ok := args.Get(0).([]byte)
		if !ok {
			suite.T().Fatal("cannot convert")
		}
		copy(b, part1)
	}).Once()
	suite.timeoutReaderMock.On("Read", mock.Anything).Return(0, os.ErrDeadlineExceeded).Once()
	part2 := expectedData[len(expectedData)/2:]
	suite.timeoutReaderMock.On("Read", mock.Anything).Return(len(part2), nil).Run(func(args mock.Arguments) {
		b, ok := args.Get(0).([]byte)
		if !ok {
			suite.T().Fatal("cannot convert")
		}
		copy(b, part2)
	}).Once()
	suite.timeoutReaderMock.On("Read", mock.Anything).Return(0, io.EOF).Maybe()

	pr := NewPromptReader(suite.timeoutReaderMock, suite.BuffZize, suite.matchLength)
	err := pr.SetPromptPattern(`(?i)user\w+:\s+`)
	suite.NoError(err)
	pr.SetDeadLine(time.Now().Add(time.Second))

	var buf bytes.Buffer
	_, err = buf.ReadFrom(pr)
	suite.NoError(err)
	suite.Equal(buf.Bytes(), expectedData)
	suite.timeoutReaderMock.AssertExpectations(suite.T())
}

func (suite *PromptReaderSuite) TestReadNoPrompt() {
	expectedData := []byte(`Hello, World! `)
	suite.timeoutReaderMock.On("Read", mock.Anything).Return(len(expectedData), nil).Run(func(args mock.Arguments) {
		b, ok := args.Get(0).([]byte)
		if !ok {
			suite.T().Fatal("cannot convert")
		}
		copy(b, expectedData)
	}).Once()
	suite.timeoutReaderMock.On("Read", mock.Anything).Return(0, io.EOF)

	pr := NewPromptReader(suite.timeoutReaderMock, suite.BuffZize, suite.matchLength)
	err := pr.SetPromptPattern(`(?i)user\w+:\s+`)
	suite.NoError(err)
	pr.SetDeadLine(time.Now().Add(50 * time.Millisecond))

	var buf bytes.Buffer
	_, err = buf.ReadFrom(pr)
	suite.ErrorIs(err, ErrNoPromptFound)
	suite.timeoutReaderMock.AssertExpectations(suite.T())
}

func (suite *PromptReaderSuite) TestReset() {
	expectedData := []byte(`Authorization required!

	Username: `)
	suite.timeoutReaderMock.On("Read", mock.Anything).Return(len(expectedData), nil).Run(func(args mock.Arguments) {
		b, ok := args.Get(0).([]byte)
		if !ok {
			suite.T().Fatal("cannot convert")
		}
		copy(b, expectedData)
	}).Once()
	// suite.timeoutReaderMock.On("Read", mock.Anything).Return(0, io.EOF).Once()

	pr := NewPromptReader(suite.timeoutReaderMock, suite.BuffZize, suite.matchLength)
	err := pr.SetPromptPattern(`(?i)user\w+:\s+`)
	suite.NoError(err)
	buffer := make([]byte, len(expectedData))
	n, err := pr.Read(buffer)
	suite.NoError(err)
	suite.Equal(n, len(expectedData))
	suite.Equal(buffer, expectedData)

	pr.Reset()

	expectedData2 := []byte(`Password: `)
	suite.timeoutReaderMock.On("Read", mock.Anything).Return(len(expectedData2), nil).Run(func(args mock.Arguments) {
		b2, ok := args.Get(0).([]byte)
		if !ok {
			suite.T().Fatal("cannot convert")
		}
		copy(b2, expectedData2)
	}).Once()
	// suite.timeoutReaderMock.On("Read", mock.Anything).Return(0, io.EOF).Once()
	err2 := pr.SetPromptPattern(`(?i)passw\w+:\s+`)
	suite.NoError(err2)
	buffer2 := make([]byte, len(expectedData2))
	n2, err2 := pr.Read(buffer2)
	suite.NoError(err2)
	suite.Equal(n2, len(expectedData2))
	suite.Equal(buffer2, expectedData2)

	suite.timeoutReaderMock.AssertExpectations(suite.T())
}

func TestPromptReaderSuite(t *testing.T) {
	suite.Run(t, new(PromptReaderSuite))
}
