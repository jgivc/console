package console

import (
	"io"
	"time"

	"golang.org/x/crypto/ssh"
)

type sshTransport struct {
	client  *ssh.Client
	session *ssh.Session
	stdin   io.Writer
	r       timeoutReader
}

func (t *sshTransport) Open(host *Host) error {
	config := &ssh.ClientConfig{
		User: host.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(host.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial(protoTCP, host.GetHostPort(), config)
	if err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}

	t.client = client
	t.session = session

	if t.stdin, err = session.StdinPipe(); err != nil {
		return err
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}

	t.r = newTimeoutReader(stdout)

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	// Request pseudo terminal
	if err := session.RequestPty("xterm", 40, 80, modes); err != nil {
		return err
	}

	err = session.Shell()
	if err != nil {
		return err
	}

	return nil
}

func (t *sshTransport) Read(b []byte) (int, error) {
	return t.r.Read(b)
}

func (t *sshTransport) Write(b []byte) (int, error) {
	return t.stdin.Write(b)
}

func (t *sshTransport) Close() error {
	if t.session != nil {
		t.session.Close()
	}

	return t.client.Close()
}

func (t *sshTransport) SetReadTimeout(d time.Duration) {
	t.r.SetTimeout(d)
}
