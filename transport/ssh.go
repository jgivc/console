package transport

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/jgivc/console/host"
	"golang.org/x/crypto/ssh"
)

type sshTransport struct {
	conn        net.Conn
	client      *ssh.Client
	session     *ssh.Session
	stdin       io.Writer
	readTimeout time.Duration
	bufSize     int
	r           timeoutReader
}

func (t *sshTransport) Open(ctx context.Context, host *host.Host) error {
	config := &ssh.ClientConfig{
		User: host.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(host.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, protoTCP, host.GetHostPort())
	if err != nil {
		return err
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, host.GetHostPort(), config)
	if err != nil {
		return err
	}

	client := ssh.NewClient(sshConn, chans, reqs)

	// client, err := ssh.Dial(protoTCP, host.GetHostPort(), config)
	// if err != nil {
	// 	return err
	// }

	session, err := client.NewSession()
	if err != nil {
		return err
	}

	t.client = client
	t.session = session
	t.conn = conn

	if t.stdin, err = session.StdinPipe(); err != nil {
		return err
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}

	t.r = newTimeoutReader(ctx, stdout, t.readTimeout, t.bufSize)

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,              // disable echoing
		ssh.TTY_OP_ISPEED: sshTtyOpISpeed, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: sshTtyOpOSpeed, // output speed = 14.4kbaud
	}
	// Request pseudo terminal
	if err2 := session.RequestPty("xterm", sshTerminalHeight, sshTerminalWidth, modes); err2 != nil {
		return err2
	}

	err = session.Shell()
	if err != nil {
		return err
	}

	return nil
}

func (t *sshTransport) Read(b []byte) (int, error) {
	return t.r.Read(b)
}

func (t *sshTransport) Write(b []byte) (int, error) {
	return t.stdin.Write(b)
}

func (t *sshTransport) Close() error {
	if t.session != nil {
		t.session.Close()
	}

	t.conn.Close()
	t.r.Close()
	return t.client.Close()
}

func (t *sshTransport) SetReadTimeout(d time.Duration) {
	t.r.SetTimeout(d)
}
