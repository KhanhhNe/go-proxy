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
	LastChecked time.Time

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

// Cleanup the resources for deallocation
type CleanupFunc func()

func NewServer(host string, port int, auth *common.ProxyAuth) *Server {
	return &Server{
		uuid.Must(uuid.NewV7()).String(),
		host,
		port,
		auth,
		30 * time.Second,
		"",
		0,
		time.Time{},
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
	return fmt.Sprintf("%s:%d - %s", s.Host, s.Port, s.Id)
}

func (s *Server) Printlnf(f string, a ...any) {
	if s.skipLogging {
		return
	}

	f = fmt.Sprintf("[ProxyServer %s:%d] ", s.Host, s.Port) + f + "\n"
	fmt.Printf(f, a...)
}

func (s *Server) Prepare() error {
	f, _, _, _ := s.getHandlers()
	return f()
}

func (s *Server) IsPrepared() bool {
	_, f, _, _ := s.getHandlers()
	return f()
}

func (s *Server) Connect(target string) (net.Conn, error) {
	_, _, f, _ := s.getHandlers()
	return f(target)
}

func (s *Server) Cleanup() {
	_, _, _, f := s.getHandlers()
	f()
}

func (s *Server) CheckServer() {
	var wg sync.WaitGroup

	start := time.Now()
	isAlive := false

	for proto := range s.Protocols {
		if proto == PROTO_Direct {
			continue
		}

		copy := NewServer(s.Host, s.Port, s.Auth)
		copy.Protocols[proto] = true
		// copy.skipLogging = true

		wg.Add(1)
		go func(p string, c *Server) {
			alive := c.CheckAlive()

			common.DataMutex.Lock()

			s.Protocols[p] = alive
			if alive {
				s.PublicIp = copy.PublicIp
				isAlive = true
			}

			common.DataMutex.Unlock()

			wg.Done()
		}(proto, copy)
	}

	wg.Wait()

	common.DataMutex.Lock()

	if isAlive {
		s.Latency = time.Since(start)
	} else {
		s.Latency = 0
	}
	s.LastChecked = time.Now()

	common.DataMutex.Unlock()

	common.DataMutex.RLock()
	protos := ""
	for proto, supported := range s.Protocols {
		if supported {
			protos += "," + proto
		}
	}
	common.DataMutex.RUnlock()
	s.Printlnf("Supported protocols: %s", strings.TrimLeft(protos, ","))
}

func (s *Server) CheckAlive() bool {
	defer s.Cleanup()

	if !s.IsPrepared() {
		err := s.Prepare()
		if err != nil {
			return false
		}
	}

	conn, err := s.Connect(common.IP_CHECK_HOST + ":80")
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

	common.DataMutex.Lock()
	s.PublicIp = string(body)
	common.DataMutex.Unlock()

	return true
}

func (s *Server) getHandlers() (PrepareFunc, IsPreparedFunc, ConnectFunc, CleanupFunc) {
	common.DataMutex.RLock()
	defer common.DataMutex.RUnlock()

	switch true {
	case s.Protocols[PROTO_Http]:
		return s.prepareHttp, s.isPreparedHttp, s.connectHttp, s.cleanupHttp
	case s.Protocols[PROTO_Socks5]:
		return s.prepareSocks5, s.isPreparedSocks5, s.connectSocks5, s.cleanupSocks5
	case s.Protocols[PROTO_Ssh]:
		return s.prepareSsh, s.isPreparedSsh, s.connectSsh, s.cleanupSsh
	case s.Protocols[PROTO_Direct]:
		return s.prepareDirect, s.isPreparedDirect, s.connectDirect, s.cleanupDirect
	}

	return func() error { return nil },
		func() bool { return false },
		func(string) (net.Conn, error) { return nil, nil },
		func() {}
}
