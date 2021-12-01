package console

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	defaultTransport = TransportTELNET
)

/**
ssh://user:pass:enablepass@host:port
telnet://user:pass:enablepass@host:port
user:pass:enablepass@host:port
user:pass:enablepass@host
user:pass@host
host
*/
type URI string

var (
	uriRegexp = regexp.MustCompile(`(?i)^(?:(?P<schema>\w+)://)?(?:(?P<user>[\w.-]+):(?P<pass>[^:]+)(?::(?P<enable>[^@]+))?@)?(?P<host>[\w.-]+)(?::(?P<port>\d{2,5}))?$`)
)

func (u URI) ToHost() (*Host, error) {
	m := uriRegexp.FindStringSubmatch(string(u))
	if m == nil {
		return nil, fmt.Errorf("cannot convert")
	}

	var h Host

	for i, name := range uriRegexp.SubexpNames() {
		switch name {
		case "schema":
			s := strings.ToLower(m[i])
			switch s {
			case "ssh":
				h.TransportType = TransportSSH
			case "telnet":
				h.TransportType = TransportTELNET
			case "":
				h.TransportType = defaultTransport
			default:
				return nil, fmt.Errorf("unknown scheme")
			}
		case "user":
			h.Account.Username = m[i]
		case "pass":
			h.Account.Password = m[i]
		case "enable":
			h.Account.EnablePassword = m[i]
		case "host":
			h.Host = m[i]
		case "port":
			if m[i] == "" {
				h.Port = -1
			} else {
				p, err := strconv.Atoi(m[i])
				if err != nil {
					return nil, err
				}

				h.Port = p
			}
		}
	}

	if h.Port < 0 {
		switch h.TransportType {
		case TransportTELNET:
			h.Port = defaultTelnetPort
		case TransportSSH:
			h.Port = defaultSSHPort
		default:
			return nil, fmt.Errorf("unknown scheme")
		}
	}

	return &h, nil
}
