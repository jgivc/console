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
