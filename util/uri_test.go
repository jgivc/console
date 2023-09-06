package util

import (
	"reflect"
	"testing"

	"github.com/jgivc/console/host"
	"github.com/jgivc/console/transport"
)

func TestURI(t *testing.T) {
	data := map[string]*host.Host{
		"10.1.1.1": {
			Host:          "10.1.1.1",
			Port:          defaultTelnetPort,
			TransportType: transport.TransportTELNET,
		},
		"ssh://10.1.1.1": {
			Host:          "10.1.1.1",
			Port:          defaultSSHPort,
			TransportType: transport.TransportSSH,
		},
		"telnet://10.1.1.1": {
			Host:          "10.1.1.1",
			Port:          defaultTelnetPort,
			TransportType: transport.TransportTELNET,
		},
		"ssh://10.1.1.1:12345": {
			Host:          "10.1.1.1",
			Port:          12345,
			TransportType: transport.TransportSSH,
		},
		"user:pass@10.1.1.1": {
			Host:          "10.1.1.1",
			Port:          defaultTelnetPort,
			TransportType: transport.TransportTELNET,
			Account: host.Account{
				Username: "user",
				Password: "pass",
			},
		},
		"user:pass:enable@10.1.1.1": {
			Host:          "10.1.1.1",
			Port:          defaultTelnetPort,
			TransportType: transport.TransportTELNET,
			Account: host.Account{
				Username:       "user",
				Password:       "pass",
				EnablePassword: "enable",
			},
		},
		"ssh://user:pass:enable@10.1.1.1": {
			Host:          "10.1.1.1",
			Port:          defaultSSHPort,
			TransportType: transport.TransportSSH,
			Account: host.Account{
				Username:       "user",
				Password:       "pass",
				EnablePassword: "enable",
			},
		},
	}

	for k := range data {
		u := URI(k)
		h, err := u.ToHost()
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(h, data[k]) {
			t.Fatalf("error for uri: '%s', want: %v but got: %v", k, data[k], h)
		}
	}
}
