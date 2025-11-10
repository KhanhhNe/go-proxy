package proxyserver

import (
	"fmt"
	"go-proxy/common"
	"net"
	"strconv"
	"time"

	"braces.dev/errtrace"
	"golang.org/x/crypto/ssh"
)

type SshServer struct {
	Host    string
	Port    int
	Timeout time.Duration
	Auth    *common.ProxyAuth

	Client *ssh.Client
}

// Enforce type check
var _ Server = &SshServer{}

func NewSshProxyServer(host string, port int, auth *common.ProxyAuth) *SshServer {
	return &SshServer{
		Host:    host,
		Port:    port,
		Auth:    auth,
		Timeout: 30 * time.Second,
	}
}

func (s *SshServer) String() string {
	return fmt.Sprintf("<%s host=%s port=%d auth=%s >", s.Type(), s.Host, s.Port, s.Auth)
}

func (s *SshServer) Type() string { return "ssh" }

func (s *SshServer) Prepare() error {
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

	s.Client = c
	return nil
}

func (s *SshServer) Prepared() bool { return s.Client != nil }

func (s *SshServer) Connect(host string) (net.Conn, error) {
	c, err := s.Client.Dial("tcp", host)
	return c, errtrace.Wrap(err)
}
