package console

import (
	"reflect"
	"testing"
)

func TestURI(t *testing.T) {
	data := map[string]*Host{
		"10.1.1.1": {
			Host:          "10.1.1.1",
			Port:          defaultTelnetPort,
			TransportType: TransportTELNET,
		},
		"ssh://10.1.1.1": {
			Host:          "10.1.1.1",
			Port:          defaultSSHPort,
			TransportType: TransportSSH,
		},
		"telnet://10.1.1.1": {
			Host:          "10.1.1.1",
			Port:          defaultTelnetPort,
			TransportType: TransportTELNET,
		},
		"ssh://10.1.1.1:12345": {
			Host:          "10.1.1.1",
			Port:          12345,
			TransportType: TransportSSH,
		},
		"user:pass@10.1.1.1": {
			Host:          "10.1.1.1",
			Port:          defaultTelnetPort,
			TransportType: TransportTELNET,
			Account: Account{
				Username: "user",
				Password: "pass",
			},
		},
		"user:pass:enable@10.1.1.1": {
			Host:          "10.1.1.1",
			Port:          defaultTelnetPort,
			TransportType: TransportTELNET,
			Account: Account{
				Username:       "user",
				Password:       "pass",
				EnablePassword: "enable",
			},
		},
		"ssh://user:pass:enable@10.1.1.1": {
			Host:          "10.1.1.1",
			Port:          defaultSSHPort,
			TransportType: TransportSSH,
			Account: Account{
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
