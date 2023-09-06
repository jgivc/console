package transport

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jgivc/console/host"
)

type ReadTimeout time.Duration

func (t *ReadTimeout) UnmarshalXMLAttr(attr xml.Attr) error {
	d, err := time.ParseDuration(attr.Value)
	if err != nil {
		return err
	}

	*t = ReadTimeout(d)

	return nil
}

type SendData struct {
	Timeout ReadTimeout `xml:"timeout,attr"`
	Send    []byte      `xml:",chardata"`
}

type ReaderData struct {
	XMLName  xml.Name   `xml:"scenario"`
	SendData []SendData `xml:"send"`
}

type dummyReader struct {
	i       int
	timeout time.Duration
	rd      ReaderData
	done    chan struct{}
}

func (r *dummyReader) Read(b []byte) (int, error) {
	if r.i >= len(r.rd.SendData) {
		return 0, io.EOF
	}

	sd := r.rd.SendData[r.i]

	if sd.Timeout < 1 {
		n := copy(b, sd.Send)
		if n != len(sd.Send) {
			return n, io.ErrShortBuffer
		}
		r.i++

		return n, io.EOF
	}

	select {
	case <-r.done:
		r.i = len(r.rd.SendData)
		return 0, fmt.Errorf("interrupted")
	case <-time.After(time.Duration(sd.Timeout)):
		n := copy(b, sd.Send)
		if n != len(sd.Send) {
			return n, io.ErrShortBuffer
		}

		r.i++
		return n, io.EOF
	case <-time.After(r.timeout):
		return 0, os.ErrDeadlineExceeded
	}
}

func (r *dummyReader) SetTimeout(timeout time.Duration) {
	r.timeout = timeout
}

func (r *dummyReader) Close() error {
	r.i = len(r.rd.SendData)
	close(r.done)

	return nil
}

type dummyTransport struct {
	fileName string
	timeout  time.Duration
	dr       *dummyReader
}

func (t *dummyTransport) Open(ctx context.Context, host *host.Host) error {
	if t.timeout < 1 {
		return fmt.Errorf("dummyTransport timeout must be set")
	}

	b, err := os.ReadFile(t.fileName)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	dr := &dummyReader{
		done:    make(chan struct{}),
		timeout: t.timeout,
	}

	if err2 := xml.Unmarshal(b, &dr.rd); err2 != nil {
		close(dr.done)
		return fmt.Errorf("cannot unmarhall data: %w", err2)
	}

	t.dr = dr
	return nil
}

func (t *dummyTransport) Read(b []byte) (int, error) {
	return t.dr.Read(b)
}

func (t *dummyTransport) Write(b []byte) (int, error) {
	return len(b), nil
}

func (t *dummyTransport) Close() error {
	if t.dr != nil {
		return t.dr.Close()
	}

	return nil
}

func (t *dummyTransport) SetReadTimeout(d time.Duration) {
	t.dr.SetTimeout(d)
}
