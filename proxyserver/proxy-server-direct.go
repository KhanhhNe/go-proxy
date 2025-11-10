package proxyserver

import (
	"fmt"
	"net"

	"braces.dev/errtrace"
)

type DirectServer struct {
}

// Enforce type check
var _ Server = &DirectServer{}

func NewDirectProxyServer() *DirectServer {
	return &DirectServer{}
}

func (s *DirectServer) String() string {
	return fmt.Sprintf("<%s host=nil port=nil auth=nil >", s.Type())
}

func (s *DirectServer) Type() string { return "direct" }

func (s *DirectServer) Prepare() error {
	return nil
}

func (s *DirectServer) Prepared() bool { return true }

func (s *DirectServer) Connect(host string) (net.Conn, error) {
	c, err := net.Dial("tcp", host)
	return c, errtrace.Wrap(err)
}
