package proxyserver

import (
	"bufio"
	"fmt"
	"go-proxy/common"
	"net"
	"net/http"
	"strings"
	"sync"
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

	skipLogging bool
}

const (
	PROTO_Ssh    = "ssh"
	PROTO_Socks5 = "socks5"
	PROTO_Http   = "http"
	PROTO_Direct = "direct"
)

// Prepare the connection, pre-authentication, etc. (e.g. connect to SSH server and authenticate)
type PrepareFunc func() error

// Check if preparation is needed before connecting
type IsPreparedFunc func() bool

// Connect and open a tunnel for 2-way connection & transfer
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

		false,
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
	if s.skipLogging {
		return
	}

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

func (s *Server) CheckProtocols() map[string]bool {
	var wg sync.WaitGroup
	res := map[string]bool{}
	var mu sync.Mutex

	for proto := range s.Protocols {
		if proto == PROTO_Direct {
			continue
		}

		copy := NewServer(s.Host, s.Port, s.Auth)
		copy.Protocols[proto] = true
		copy.skipLogging = true

		wg.Add(1)
		go func(p string, c *Server) {
			mu.Lock()
			res[p] = c.CheckAlive()
			mu.Unlock()
			wg.Done()
		}(proto, copy)
	}

	wg.Wait()

	protos := ""
	for proto, supported := range res {
		if supported {
			protos += "," + proto
		}
	}
	s.Printlnf("Supported protocols: %s", strings.TrimLeft(protos, ","))

	return res
}

func (s *Server) CheckAlive() bool {
	prepare, isPrepared, connect := s.getHandlers()

	if !isPrepared() {
		err := prepare()
		if err != nil {
			return false
		}
	}

	conn, err := connect(common.IP_CHECK_HOST + ":80")
	if err != nil || conn == nil {
		return false
	}
	defer conn.Close()

	req, err := http.NewRequest("GET", "http://"+common.IP_CHECK_HOST, nil)
	if err != nil {
		fmt.Printf("Unexpected error while requesting in CheckAlive: %+v", errtrace.Wrap(err))
		return false
	}

	err = req.Write(conn)
	if err != nil {
		return false
	}

	res, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false
	}

	return strings.Contains(res, "HTTP")
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
