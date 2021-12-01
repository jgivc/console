package console

import (
	"bytes"
	"fmt"
	"io"
	"time"
)

const (
	defaultTimeout = 30 * time.Millisecond
)

var (
	errorRreadTimeout = fmt.Errorf("read timeout")
)

type timeoutReader interface {
	io.ReadCloser
	SetTimeout(t time.Duration)
}

type rd struct {
	b   []byte
	err error
}

type tReader struct {
	r      io.Reader
	t      time.Duration
	ch     chan rd
	buf    bytes.Buffer
	closed bool
}

func (r *tReader) Read(b []byte) (int, error) {
	if r.closed {
		return 0, fmt.Errorf("read on closed reader")
	}

	if r.buf.Len() < 1 {
		select {
		case bb := <-r.ch:
			if bb.err != nil {
				return 0, bb.err
			}
			if err := sCopy(&r.buf, bb.b); err != nil {
				return 0, err
			}
		case <-time.After(r.t):
			return 0, errorRreadTimeout
		}

		if r.closed {
			return 0, fmt.Errorf("channel closed")
		}

	}

	return r.buf.Read(b)
}

func (r *tReader) Close() error {
	r.closed = true
	return nil
}

func (r *tReader) SetTimeout(t time.Duration) {
	r.t = t
}

func newTimeoutReader(r io.Reader) timeoutReader {
	ch := make(chan rd)

	go func() {

		b := make([]byte, 1024)

		ch <- rd{}

		for {
			n, err := r.Read(b)
			if err != nil {
				ch <- rd{nil, err}
				break
			}

			tmp := make([]byte, n)
			copy(tmp, b[:n])

			ch <- rd{tmp, nil}
		}

	}()

	<-ch

	return &tReader{
		r:  r,
		t:  defaultTimeout,
		ch: ch,
	}
}

func sCopy(w io.Writer, b []byte) error {
	for {
		n, err := w.Write(b)
		if err != nil {
			return err
		}
		if n < len(b) {
			b = b[n:]
		} else {
			break
		}
	}

	return nil
}
