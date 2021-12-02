package console

import (
	"strings"
	"time"
)

const (
	promptWaitTimeout     = 10 * time.Second
	promptMatchLen    int = 15
	readBufferSize        = 200

	authPattern   = `(?mi)(user\w+|pass\w+|[\w-()\s]+)[#>:]+(?:\s+)?`
	promptPattern = `(?mi)([\w-()\s]{2,})[#>]+(?:\s+)?\z`

	userPromptPart     = "user"
	passwordPromptPart = "pass"
)

var (
	cmdEnd = []byte("\r")
)

//Console is uniform interface for interacting with network hardware via telnet/ssh
type Console interface {
	Open(host *Host) error
	Execute(cmd string) (string, error)
	Run(cmd string) error // Just run command and read and omit output
	Send(cmd string) error
	Sendln(cmd string) error
	SetPrompt(pattern string)
	Close() error
}

type console struct {
	h   *Host
	tr  transport
	pm  promptMatcher
	buf []byte
}

func (c *console) tryAuth() error {
	c.pm = newPromptRegexpMatcher(authPattern)
	for {
		_, err := c.readToPrompt()
		if err != nil {
			return err
		}

		m := c.pm.getMatched()
		if m != nil {
			if ss, ok := m.([]string); ok {
				line := ss[1]
				line = strings.ToLower(line)
				if strings.Contains(line, userPromptPart) {
					c.Sendln(c.h.Username)
				} else if strings.Contains(line, passwordPromptPart) {
					c.Sendln(c.h.Password)
				} else {
					break
				}
			}
		}
	}

	c.pm = newPromptRegexpMatcher(promptPattern)
	return nil
}

func (c *console) readToPrompt() (string, error) {
	b := make([]byte, 0)
	start := time.Now()

	for {
		n, err := c.tr.Read(c.buf)
		if err != nil {
			if err == errorRreadTimeout {
				if time.Since(start) < promptWaitTimeout {
					l := len(b)
					if l > promptMatchLen {
						if c.pm.match(string(b[l-promptMatchLen:])) {
							break
						}
					} else {
						if c.pm.match(string(b)) {
							break
						}
					}
					continue
				}

				return "", err
			}
		}

		b = append(b, c.buf[:n]...)
		if l := len(b); l > promptMatchLen {
			if c.pm.match(string(b[l-promptMatchLen:])) {
				break
			}
		}
	}

	return string(b), nil
}

func (c *console) Open(host *Host) error {
	var err error
	c.tr, err = newTransport(host.TransportType)
	if err != nil {
		return err
	}

	if err := c.tr.Open(host); err != nil {
		return err
	}

	c.h = host

	if err := c.tryAuth(); err != nil {
		return err
	}

	return nil

}

func (c *console) Execute(cmd string) (string, error) {
	c.Sendln(cmd)
	return c.readToPrompt()
}

func (c *console) Run(cmd string) error {
	c.Sendln(cmd)
	_, err := c.readToPrompt()

	return err
}

func (c *console) SetPrompt(pattern string) {
	c.pm = newPromptRegexpMatcher(pattern)
}

func (c *console) Close() error {
	return c.tr.Close()
}

func (c *console) Send(cmd string) error {
	c.tr.Write([]byte(cmd))

	return nil
}

func (c *console) Sendln(cmd string) error {
	c.tr.Write([]byte(cmd))
	c.tr.Write(cmdEnd)

	return nil
}

//New function create Console instance
func New() Console {
	return &console{
		buf: make([]byte, readBufferSize),
	}
}
