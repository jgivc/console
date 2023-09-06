package transport

import (
	"context"
	"io"
	"time"

	"github.com/jgivc/console/host"
)

type timeoutReader interface {
	io.ReadCloser
	SetTimeout(t time.Duration)
}

type Transport interface {
	Open(ctx context.Context, host *host.Host) error
	SetReadTimeout(t time.Duration)
	io.ReadWriteCloser
}
