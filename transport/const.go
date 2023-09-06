package transport

const (
	protoTCP = "tcp"

	TransportSSH = iota
	TransportTELNET
	TransportDummy

	sshTtyOpISpeed    = 115200
	sshTtyOpOSpeed    = 115200
	sshTerminalWidth  = 80
	sshTerminalHeight = 1000
)
