package proxyserver

import (
	"net"
	"strconv"

	"braces.dev/errtrace"
	"golang.org/x/crypto/ssh"
)

type ServerSshState struct {
	client *ssh.Client
}

func (s *Server) prepareSsh() error {
	c, err := ssh.Dial("tcp", net.JoinHostPort(s.Host, strconv.Itoa(s.Port)), &ssh.ClientConfig{
		User: s.Auth.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(s.Auth.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         s.Timeout,
	})
	if err != nil {
		return errtrace.Wrap(err)
	}

	s.sshState.client = c
	return nil
}

func (s *Server) isPreparedSsh() bool { return s.sshState.client != nil }

func (s *Server) connectSsh(target string) (net.Conn, error) {
	c, err := s.sshState.client.Dial("tcp", target)
	return c, errtrace.Wrap(err)
}
