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

/*URI represent host connection string
Example:

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

//ToHost method convert connection string ro *Host instance
func (u URI) ToHost() (*Host, error) {
	m := uriRegexp.FindStringSubmatch(string(u))
	if m == nil {
		return nil, fmt.Errorf("cannot convert")
	}

	var h Host
	var strPort string

	for i, name := range uriRegexp.SubexpNames() {
		switch name {
		case "schema":
			tt, err := getTransportType(m[i])
			if err != nil {
				return nil, err
			}

			h.TransportType = tt
		case "user":
			h.Account.Username = m[i]
		case "pass":
			h.Account.Password = m[i]
		case "enable":
			h.Account.EnablePassword = m[i]
		case "host":
			h.Host = m[i]
		case "port":
			strPort = m[i]
		}
	}

	port, err := getPort(strPort, h.TransportType)
	if err != nil {
		return nil, err
	}

	h.Port = port

	return &h, nil
}

func getPort(s string, tt int) (port int, err error) {
	if s == "" {
		port = -1
	} else {
		port, err = strconv.Atoi(s)
		if err != nil {
			return
		}
	}

	if port < 0 {
		switch tt {
		case TransportTELNET:
			port = defaultTelnetPort
		case TransportSSH:
			port = defaultSSHPort
		default:
			err = fmt.Errorf("unknown scheme")
		}
	}

	return
}

func getTransportType(s string) (tt int, err error) {
	s = strings.ToLower(s)
	switch s {
	case "ssh":
		tt = TransportSSH
	case "telnet":
		tt = TransportTELNET
	case "":
		tt = defaultTransport
	default:
		err = fmt.Errorf("unknown scheme")
	}

	return
}
