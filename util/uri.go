package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jgivc/console/host"
	"github.com/jgivc/console/transport"
)

const (
	defaultSSHPort    = 22
	defaultTelnetPort = 23
	defaultTransport  = transport.TransportTELNET
)

/*
URI represent host connection string
Example:

ssh://user:pass:enablepass@host:port
telnet://user:pass:enablepass@host:port
user:pass:enablepass@host:port
user:pass:enablepass@host
user:pass@host
host.
*/
type URI string

var (
	uriRegexp = regexp.MustCompile(`(?i)^(?:(?P<schema>\w+)://)?(?:(?P<user>[\w.-]+):(?P<pass>[^:]+)` +
		`(?::(?P<enable>[^@]+))?@)?(?P<host>[\w.-]+)(?::(?P<port>\d{2,5}))?$`)
)

// ToHost method convert connection string ro *Host instance.
func (u URI) ToHost() (*host.Host, error) {
	m := uriRegexp.FindStringSubmatch(string(u))
	if m == nil {
		return nil, fmt.Errorf("cannot convert")
	}

	var h host.Host
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

func getPort(s string, tt int) (int, error) {
	var (
		port int
		err  error
	)

	if s == "" {
		port = -1
	} else {
		port, err = strconv.Atoi(s)
		if err != nil {
			return port, err
		}
	}

	if port < 0 {
		switch tt {
		case transport.TransportTELNET:
			port = defaultTelnetPort
		case transport.TransportSSH:
			port = defaultSSHPort
		default:
			return -1, fmt.Errorf("unknown scheme")
		}
	}

	return port, nil
}

func getTransportType(s string) (int, error) {
	var tt int

	s = strings.ToLower(s)
	switch s {
	case "ssh":
		tt = transport.TransportSSH
	case "telnet":
		tt = transport.TransportTELNET
	case "":
		tt = defaultTransport
	default:
		return -1, fmt.Errorf("unknown scheme")
	}

	return tt, nil
}

type HostFactory interface {
	GetHost(uri string) (*host.Host, error)
}

type hf struct {
	account host.Account
}

func (f *hf) GetHost(uri string) (*host.Host, error) {
	host, err := URI(uri).ToHost()
	if err != nil {
		return nil, err
	}

	if !host.HasAccount() {
		host.Account = f.account
	}

	return host, nil
}

func NewHostFactory(account host.Account) HostFactory {
	return &hf{account: account}
}
