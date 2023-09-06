package transport

import (
	"context"
	"time"

	"github.com/jgivc/console/host"
	"github.com/jgivc/console/transport/telnet"
)

type telnetTransport struct {
	conn        *telnet.Conn
	readTimeout time.Duration
	bufSize     int
	r           timeoutReader
}

func (t *telnetTransport) Open(ctx context.Context, host *host.Host) error {
	conn, err := telnet.DialContext(ctx, host.GetHostPort())
	if err != nil {
		return err
	}

	t.conn = conn
	t.r = newTimeoutReader(ctx, conn, t.readTimeout, t.bufSize)

	return nil
}

func (t *telnetTransport) Read(b []byte) (int, error) {
	return t.r.Read(b)
}

func (t *telnetTransport) Write(b []byte) (int, error) {
	return t.conn.Write(b)
}

func (t *telnetTransport) Close() error {
	var err error

	if t.conn != nil {
		t.r.Close()
		err = t.conn.Close()
		t.conn = nil
	}

	return err
}

func (t *telnetTransport) SetReadTimeout(d time.Duration) {
	t.r.SetTimeout(d)
}
