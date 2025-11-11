package proxyserver

import (
	"fmt"
	"go-proxy/common"
	"net"
	"strings"
	"time"

	"braces.dev/errtrace"
)

type Server struct {
	Host    string
	Port    int
	Auth    *common.ProxyAuth
	Timeout time.Duration

	Protocols map[string]bool

	// Protocol-specific state
	sshState    *ServerSshState
	httpState   *ServerHttpState
	socks5State *ServerSocks5State
	directState *ServerDirectState
}

const (
	PROTO_Ssh    = "ssh"
	PROTO_Socks5 = "socks5"
	PROTO_Http   = "http"
	PROTO_Direct = "direct"
)

type PrepareFunc func() error
type IsPreparedFunc func() bool
type ConnectFunc func(string) (net.Conn, error)

func NewServer(host string, port int, auth *common.ProxyAuth) *Server {
	return &Server{
		host,
		port,
		auth,
		30 * time.Second,
		map[string]bool{
			PROTO_Ssh:    false,
			PROTO_Socks5: false,
			PROTO_Http:   false,
			PROTO_Direct: false,
		},

		&ServerSshState{},
		&ServerHttpState{},
		&ServerSocks5State{},
		&ServerDirectState{},
	}
}

func (s *Server) String() string {
	protos := ""
	if s.Protocols[PROTO_Http] {
		protos += ",http"
	}
	if s.Protocols[PROTO_Socks5] {
		protos += ",socks5"
	}
	if s.Protocols[PROTO_Ssh] {
		protos += ",ssh"
	}
	protos = strings.TrimLeft(protos, ",")
	if protos == "" {
		protos = "no_proto"
	}

	return fmt.Sprintf("%s %s:%d", protos, s.Host, s.Port)
}

func (s *Server) Printlnf(f string, a ...any) {
	f = fmt.Sprintf("[ProxyServer %s] ", s.String()) + f + "\n"
	fmt.Printf(f, a...)
}

func (s *Server) Prepare() error {
	f, _, _ := s.getHandlers()
	return f()
}

func (s *Server) IsPrepared() bool {
	_, f, _ := s.getHandlers()
	return f()
}

func (s *Server) Connect(target string) (net.Conn, error) {
	_, _, f := s.getHandlers()
	return f(target)
}

func (s *Server) getHandlers() (PrepareFunc, IsPreparedFunc, ConnectFunc) {
	switch true {
	case s.Protocols[PROTO_Http]:
		return s.prepareHttp, s.isPreparedHttp, s.connectHttp
	case s.Protocols[PROTO_Socks5]:
		return s.prepareSocks5, s.isPreparedSocks5, s.connectSocks5
	case s.Protocols[PROTO_Ssh]:
		return s.prepareSsh, s.isPreparedSsh, s.connectSsh
	case s.Protocols[PROTO_Direct]:
		return s.prepareDirect, s.isPreparedDirect, s.connectDirect
	}

	panic(errtrace.Errorf("No supported protocol for this server"))
}
