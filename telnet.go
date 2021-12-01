package console

import (
	"time"

	"github.com/jgivc/console/telnet"
)

type telnetTransport struct {
	conn *telnet.Conn
	r    timeoutReader
}

func (t *telnetTransport) Open(host *Host) error {

	conn, err := telnet.DialTo(host.GetHostPort())
	if err != nil {
		return err
	}

	t.conn = conn
	t.r = newTimeoutReader(conn)

	return nil
}

func (t *telnetTransport) Read(b []byte) (int, error) {
	return t.r.Read(b)
}

func (t *telnetTransport) Write(b []byte) (int, error) {
	return t.conn.Write(b)
}

func (t *telnetTransport) Close() error {
	return t.conn.Close()
}

func (t *telnetTransport) SetReadTimeout(d time.Duration) {
	t.r.SetTimeout(d)
}
