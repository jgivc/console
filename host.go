package console

import "fmt"

type Account struct {
	Username, Password, EnablePassword string
}

type Host struct {
	Host          string
	Port          int
	TransportType int
	Account
}

func (h *Host) GetHostPort() string {
	return fmt.Sprintf("%s:%d", h.Host, h.Port)
}

func (h *Host) HasAccount() bool {
	return !(h.Account == (Account{}))
}
