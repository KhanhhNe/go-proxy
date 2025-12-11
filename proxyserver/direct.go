package proxyserver

import (
	"net"

	"braces.dev/errtrace"
)

type ServerDirectState struct {
}

func NewDirectServer() *Server {
	s := NewServer("127.0.0.1", 0, nil)
	s.Protocols[PROTO_Direct] = true
	return s
}

func (s *Server) prepareDirect() error {
	return nil // No preparation needed
}

func (s *Server) isPreparedDirect() bool { return true }

func (s *Server) connectDirect(target string) (net.Conn, error) {
	c, err := net.Dial("tcp", target)
	return c, errtrace.Wrap(err)
}

func (s *Server) cleanupDirect() {}
