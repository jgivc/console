package transport

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"
)

func newTimeoutReader(ctx context.Context, reader io.Reader, timeout time.Duration, bufSize int) timeoutReader {
	ctx2, cancel := context.WithCancel(ctx)

	tr := &timeoutReaderImpl{
		reader:  reader,
		timeout: timeout,
		buf:     make([]byte, bufSize),
		chData:  make(chan []byte),
		chErr:   make(chan error, 1),
		cancel:  cancel,
	}

	go tr.readLoop(ctx2)

	return tr
}

type timeoutReaderImpl struct {
	reader   io.Reader
	timeout  time.Duration
	buf      []byte
	chData   chan ([]byte)
	chErr    chan error
	cancel   context.CancelFunc
	closed   atomic.Bool
	restData []byte
}

func (tr *timeoutReaderImpl) readLoop(ctx context.Context) {
	defer func() {
		close(tr.chErr)
		close(tr.chData)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := tr.reader.Read(tr.buf)
			if err != nil {
				tr.chErr <- err
				return
			}
			if n > 0 {
				data := make([]byte, n)
				n2 := copy(data, tr.buf[:n])
				if n != n2 {
					tr.chErr <- io.ErrShortBuffer
				}
				tr.chData <- data
			}
		}
	}
}

func (tr *timeoutReaderImpl) Read(p []byte) (int, error) {
	if tr.restData != nil {
		n := copy(p, tr.restData)
		if n < len(tr.restData) {
			tr.restData = tr.restData[n:]
		} else {
			tr.restData = nil
		}

		return n, nil
	}

	select {
	case data, ok := <-tr.chData:
		if !ok {
			return 0, io.EOF
		}
		n := copy(p, data)
		if n < len(data) {
			tr.restData = data[n:]
		}
		return n, nil
	case err := <-tr.chErr:
		return 0, fmt.Errorf("read error: %w", err)
	case <-time.After(tr.timeout):
		return 0, os.ErrDeadlineExceeded
	}
}

func (tr *timeoutReaderImpl) SetTimeout(timeout time.Duration) {
	tr.timeout = timeout
}

func (tr *timeoutReaderImpl) Close() error {
	if !tr.closed.Load() {
		tr.cancel()
		tr.closed.Store(true)
	}

	return nil
}
