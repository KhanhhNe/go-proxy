package proxyserver

import (
	"errors"
	"io"
	"net"
	"strconv"

	"braces.dev/errtrace"
	"golang.org/x/crypto/ssh"
)

type ServerSshState struct {
	client *ssh.Client
}

func (s *Server) prepareSsh() error {
	s.Printlnf("Connecting to remote server")
	c, err := ssh.Dial("tcp", net.JoinHostPort(s.Host, strconv.Itoa(s.Port)), &ssh.ClientConfig{
		User: s.Auth.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(s.Auth.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         s.Timeout,
	})
	if err != nil {
		err = errtrace.Wrap(err)
		s.Printlnf("Connect to remote server failed. Error: %+v", err)
		return err
	}
	s.Printlnf("Connection succeeded")

	s.sshState.client = c
	return nil
}

func (s *Server) isPreparedSsh() bool { return s.sshState.client != nil }

func (s *Server) connectSsh(target string) (net.Conn, error) {
	c, err := s.connectSshRetry(target, 1)
	return c, errtrace.Wrap(err)
}

func (s *Server) connectSshRetry(target string, retries int) (net.Conn, error) {
	c, err := s.sshState.client.Dial("tcp", target)

	if errors.Is(err, io.EOF) && retries > 0 {
		// EOF means connection closed from remote side
		// We'll try to connect again here
		s.Printlnf("Preparing connection again before retrying")
		er := s.prepareSsh()
		if er != nil {
			return nil, errtrace.Wrap(er)
		}

		return s.connectSshRetry(target, retries-1)
	}
	return c, errtrace.Wrap(err)
}
