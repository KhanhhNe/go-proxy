package proxyserver

import (
	"bufio"
	"fmt"
	"go-proxy/common"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"braces.dev/errtrace"
	"github.com/google/uuid"
)

type Server struct {
	Id          string
	Host        string
	Port        int
	Auth        *common.ProxyAuth
	Timeout     time.Duration
	PublicIp    string
	Latency     time.Duration
	LastChecked *time.Time

	Protocols map[string]bool
	Mu        sync.Mutex

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
		uuid.NewString(),
		host,
		port,
		auth,
		30 * time.Second,
		"",
		0,
		nil,
		map[string]bool{
			PROTO_Ssh:    false,
			PROTO_Socks5: false,
			PROTO_Http:   false,
			PROTO_Direct: false,
		},
		sync.Mutex{},

		&ServerSshState{},
		&ServerHttpState{},
		&ServerSocks5State{},
		&ServerDirectState{},

		false,
	}
}

func (s *Server) String() string {
	return fmt.Sprintf("%s:%d - %s", s.Host, s.Port, s.Id)
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

func (s *Server) CheckServer() {
	var wg sync.WaitGroup

	start := time.Now()

	for proto := range s.Protocols {
		if proto == PROTO_Direct {
			continue
		}

		copy := NewServer(s.Host, s.Port, s.Auth)
		copy.Protocols[proto] = true
		copy.skipLogging = true

		wg.Add(1)
		go func(p string, c *Server) {
			alive := c.CheckAlive()
			s.Mu.Lock()
			s.Protocols[p] = alive
			if alive {
				s.PublicIp = copy.PublicIp
			}
			s.Mu.Unlock()
			wg.Done()
		}(proto, copy)
	}

	wg.Wait()

	s.Mu.Lock()

	if s.LastChecked == nil {
		s.LastChecked = new(time.Time)
	}

	s.Latency = time.Since(start)
	*s.LastChecked = time.Now()

	s.Mu.Unlock()

	protos := ""
	for proto, supported := range s.Protocols {
		if supported {
			protos += "," + proto
		}
	}
	s.Printlnf("Supported protocols: %s", strings.TrimLeft(protos, ","))
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

	res, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return false
	}

	body, err := io.ReadAll(res.Body)
	if err != nil && err != io.EOF {
		return false
	}

	s.Mu.Lock()
	s.PublicIp = string(body)
	s.Mu.Unlock()

	return true
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
