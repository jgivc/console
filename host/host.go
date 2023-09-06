package host

import "fmt"

type Account struct {
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	EnablePassword string `yaml:"enable_password"`
}

type Host struct {
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
	TransportType int    `yaml:"transport_type"`
	Account       `yaml:",inline"`
}

func (h *Host) GetHostPort() string {
	return fmt.Sprintf("%s:%d", h.Host, h.Port)
}

func (h *Host) HasAccount() bool {
	return !(h.Account == (Account{}))
}
