package console

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jgivc/console/config"
	"github.com/jgivc/console/host"
	"github.com/jgivc/console/transport"
	"github.com/jgivc/console/util"
)

var (
	cmdEnd = []byte("\r")
)

type TransportFactory interface {
	GetTransport(host *host.Host) (transport.Transport, error)
}

type promptReader interface {
	SetPromptPattern(pattern string) error
	SetDeadLine(deadLine time.Time)
	Reset()
	io.Reader
}

type Console interface {
	Open(ctx context.Context, host *host.Host) error
	Execute(cmd string) (string, error)
	GetCommandResultReader(cmd string) (io.Reader, error)
	Run(cmd string) error // Just run command and read and omit output
	Send(cmd string) error
	Sendln(cmd string) error
	SetPrompt(pattern string) error
	Close() error
}

type console struct {
	host         *host.Host
	factory      TransportFactory
	transport    transport.Transport
	promptReader promptReader
	cfg          *config.ConsoleConfig
}

func (c *console) tryAuth() error {
	if err := c.promptReader.SetPromptPattern(c.cfg.AuthPromptPattern); err != nil {
		return fmt.Errorf("cannot set authPromptPattern: %w", err)
	}
	c.promptReader.SetDeadLine(time.Now().Add(c.cfg.AuthTimeout))

	var enable bool

	var buf bytes.Buffer
	for {
		_, err := buf.ReadFrom(c.promptReader)
		if err != nil {
			return fmt.Errorf("auth fail: %w", err)
		}

		if strings.Contains(strings.ToLower(buf.String()), c.cfg.UsernamePromptContains) {
			if err2 := c.Sendln(c.host.Username); err2 != nil {
				return fmt.Errorf("auth fail: %w", err2)
			}
		} else if strings.Contains(strings.ToLower(buf.String()), c.cfg.PasswordPromptContains) {
			if enable {
				if err2 := c.Sendln(c.host.EnablePassword); err2 != nil {
					return fmt.Errorf("auth fail: %w", err2)
				}
			} else {
				if err2 := c.Sendln(c.host.Password); err2 != nil {
					return fmt.Errorf("auth fail: %w", err2)
				}
			}
		} else if strings.HasSuffix(strings.TrimSpace(buf.String()), c.cfg.PromptSuffix) {
			if err2 := c.promptReader.SetPromptPattern(c.cfg.PromptPattern); err2 != nil {
				return fmt.Errorf("cannot set promptPattern: %w", err2)
			}
			return nil
		} else if strings.HasSuffix(strings.TrimSpace(buf.String()), c.cfg.EnableSuffix) {
			enable = true
			if err2 := c.Sendln(c.cfg.EnableCommand); err2 != nil {
				return fmt.Errorf("auth fail: %w", err2)
			}
		} else {
			return fmt.Errorf("cannot login")
		}

		c.promptReader.Reset()
		buf.Reset()
	}
}

func (c *console) Open(ctx context.Context, host *host.Host) error {
	var err error
	c.transport, err = c.factory.GetTransport(host)
	if err != nil {
		return err
	}

	if err2 := c.transport.Open(ctx, host); err2 != nil {
		return err2
	}

	c.host = host
	c.promptReader = util.NewPromptReader(c.transport, c.cfg.TransportReaderBufferSize, c.cfg.PromptMatchLengt)

	return c.tryAuth()
}

func (c *console) Execute(cmd string) (string, error) {
	c.promptReader.Reset()
	if err := c.Sendln(cmd); err != nil {
		return "", fmt.Errorf("cannot execute cmd: %w", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(c.promptReader); err != nil {
		return "", fmt.Errorf("cannot execute cmd: %w", err)
	}

	return buf.String(), nil
}

func (c *console) GetCommandResultReader(cmd string) (io.Reader, error) {
	c.promptReader.Reset()
	if err := c.Sendln(cmd); err != nil {
		return nil, fmt.Errorf("cannot execute cmd: %w", err)
	}

	return c.promptReader, nil
}

func (c *console) Run(cmd string) error {
	c.promptReader.Reset()
	if err := c.Sendln(cmd); err != nil {
		return fmt.Errorf("cannot execute cmd: %w", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(c.promptReader); err != nil {
		return fmt.Errorf("cannot execute cmd: %w", err)
	}

	return nil
}

func (c *console) Send(cmd string) error {
	_, err := c.transport.Write([]byte(cmd))

	return err
}

func (c *console) Sendln(cmd string) error {
	if _, err := c.transport.Write([]byte(cmd)); err != nil {
		return err
	}

	_, err := c.transport.Write(cmdEnd)

	return err
}

func (c *console) SetPrompt(pattern string) error {
	return c.promptReader.SetPromptPattern(pattern)
}

func (c *console) Close() error {
	return c.transport.Close()
}

func New() Console {
	return NewWithConfig(config.DefaultConsoleConfig())
}

func NewWithConfig(cfg *config.ConsoleConfig) Console {
	return &console{
		cfg: cfg,
		factory: &transport.Factory{
			DummyFileName: cfg.DummyTransportFileName,
			ReadTimeout:   cfg.TransportReadTimeout,
			BufSize:       cfg.TransportReaderBufferSize,
		},
	}
}
