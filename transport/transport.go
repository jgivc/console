package transport

import (
	"fmt"
	"time"

	"github.com/jgivc/console/host"
)

func New(t int, readTimeout time.Duration, bufSize int, dummyFileName string) (Transport, error) {
	switch t {
	case TransportSSH:
		return &sshTransport{
			readTimeout: readTimeout,
			bufSize:     bufSize,
		}, nil
	case TransportTELNET:
		return &telnetTransport{
			readTimeout: readTimeout,
			bufSize:     bufSize,
		}, nil
	case TransportDummy:
		return &dummyTransport{
			fileName: dummyFileName,
			timeout:  readTimeout,
		}, nil
	}

	return nil, fmt.Errorf("unknown transport type")
}

type Factory struct {
	DummyFileName string
	ReadTimeout   time.Duration
	BufSize       int
}

func (f *Factory) GetTransport(host *host.Host) (Transport, error) {
	return New(host.TransportType, f.ReadTimeout, f.BufSize, f.DummyFileName)
}
