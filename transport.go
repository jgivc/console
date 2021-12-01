package console

import (
	"fmt"
	"io"
	"time"
)

const (
	protoTCP = "tcp"

	defaultSSHPort    = 22
	defaultTelnetPort = 23

	//TransportSSH transport used ssh connection
	TransportSSH = iota
	//TransportTELNET transport used telnet connection
	TransportTELNET
)

type transport interface {
	Open(host *Host) error
	SetReadTimeout(t time.Duration)
	io.ReadWriteCloser
}

func newTransport(t int) (transport, error) {
	switch t {
	case TransportSSH:
		return &sshTransport{}, nil
	case TransportTELNET:
		return &telnetTransport{}, nil
	}

	return nil, fmt.Errorf("unknown transport type")
}
