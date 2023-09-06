package console

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jgivc/console/config"
	"github.com/jgivc/console/host"
	"github.com/jgivc/console/transport"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockTransport struct {
	mock.Mock
}

func (m *MockTransport) Open(ctx context.Context, host *host.Host) error {
	args := m.Called(ctx, host)
	return args.Error(0)
}

func (m *MockTransport) SetReadTimeout(t time.Duration) {
	m.Called(t)
}

func (m *MockTransport) Read(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockTransport) Write(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockTransport) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockTransportFactory struct {
	mock.Mock
}

func (m *MockTransportFactory) GetTransport(host *host.Host) (transport.Transport, error) {
	args := m.Called(host)

	var tr transport.Transport
	if args.Get(0) != nil {
		tr = args.Get(0).(transport.Transport)
	}

	return tr, args.Error(1)
}

type ConsoleTestSuite struct {
	suite.Suite
	console   *console
	factory   *MockTransportFactory
	transport *MockTransport
}

func (suite *ConsoleTestSuite) SetupTest() {
	suite.factory = new(MockTransportFactory)
	suite.transport = new(MockTransport)
	suite.console = &console{
		cfg:     config.DefaultConsoleConfig(),
		factory: suite.factory,
	}
}

func (suite *ConsoleTestSuite) TestOpenSuccess() {
	suite.transport.On("Open", mock.Anything, mock.Anything).Return(nil)
	promptLogin := []byte(`Authentication required!
	
	Username: `)
	suite.transport.On("Read", mock.Anything).Return(len(promptLogin), nil).Run(func(args mock.Arguments) {
		b, ok := args.Get(0).([]byte)
		if !ok {
			suite.T().Fatal("cannot convert")
		}
		copy(b, promptLogin)
	}).Once()
	suite.transport.On("Read", mock.Anything).Return(0, os.ErrDeadlineExceeded).Once()

	promptPassword := []byte("Password: ")
	suite.transport.On("Read", mock.Anything).Return(len(promptPassword), nil).Run(func(args mock.Arguments) {
		b, ok := args.Get(0).([]byte)
		if !ok {
			suite.T().Fatal("cannot convert")
		}
		copy(b, promptPassword)
	}).Once()
	suite.transport.On("Read", mock.Anything).Return(0, os.ErrDeadlineExceeded).Once()

	prompt := "sw1#"
	suite.transport.On("Read", mock.Anything).Return(len(prompt), nil).Run(func(args mock.Arguments) {
		b, ok := args.Get(0).([]byte)
		if !ok {
			suite.T().Fatal("cannot convert")
		}
		copy(b, prompt)
	}).Once()
	suite.transport.On("Read", mock.Anything).Return(0, os.ErrDeadlineExceeded).Maybe()

	username := "user125"
	password := "p@ssw0rD!"
	var expectedUsername, expectedPassword string

	suite.transport.On("Write", mock.Anything).Return(len(username), nil).Run(func(args mock.Arguments) {
		b, ok := args.Get(0).([]byte)
		if !ok {
			suite.T().Fatal("cannot convert")
		}

		expectedUsername = string(b)
	}).Once()
	suite.transport.On("Write", mock.Anything).Return(1, nil).Once() // '\r'
	suite.transport.On("Write", mock.Anything).Return(len(password), nil).Run(func(args mock.Arguments) {
		b, ok := args.Get(0).([]byte)
		if !ok {
			suite.T().Fatal("cannot convert")
		}

		expectedPassword = string(b)
	}).Once()
	suite.transport.On("Write", mock.Anything).Return(1, nil).Once() // '\r'

	suite.factory.On("GetTransport", mock.Anything).Return(suite.transport, nil)
	err := suite.console.Open(context.Background(), &host.Host{
		Account: host.Account{
			Username: username,
			Password: password,
		},
	})
	suite.NoError(err)
	suite.Equal(expectedUsername, username)
	suite.Equal(expectedPassword, password)

	suite.factory.AssertExpectations(suite.T())
	suite.transport.AssertExpectations(suite.T())
}

func TestConsoleTestSuite(t *testing.T) {
	suite.Run(t, new(ConsoleTestSuite))
}
