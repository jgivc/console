package console

import "fmt"

//Account struct contains username, password and enable password (if needed)
type Account struct {
	Username, Password, EnablePassword string
}

//Host struct represent host to which connect.
type Host struct {
	Host          string
	Port          int
	TransportType int
	Account
}

//GetHostPort method return connect string host:port
func (h *Host) GetHostPort() string {
	return fmt.Sprintf("%s:%d", h.Host, h.Port)
}

//HasAccount method shows whether host has an account or not
func (h *Host) HasAccount() bool {
	return !(h.Account == (Account{}))
}

//HostFactory convert uri to host and set default Account if not exist
type HostFactory interface {
	GetHost(uri string) (*Host, error)
}

type hf struct {
	account Account
}

//GetHost return *Host from string uri
func (f *hf) GetHost(uri string) (host *Host, err error) {
	host, err = URI(uri).ToHost()
	if err != nil {
		return
	}

	if !host.HasAccount() {
		host.Account = f.account
	}

	return
}

func NewHostFactory(account Account) HostFactory {
	return &hf{account: account}
}
