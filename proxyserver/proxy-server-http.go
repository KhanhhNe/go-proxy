package proxyserver

import (
	"bufio"
	"fmt"
	"go-proxy/common"
	"net"
	"net/http"
	"strconv"
	"time"

	"braces.dev/errtrace"
)

type HttpServer struct {
	Host    string
	Port    int
	Timeout time.Duration
	Auth    *common.ProxyAuth
}

// Enforce type check
var _ Server = &HttpServer{}

func NewHttpProxyServer(host string, port int, auth *common.ProxyAuth) *HttpServer {
	return &HttpServer{
		Host:    host,
		Port:    port,
		Auth:    auth,
		Timeout: 30 * time.Second,
	}
}

func (s *HttpServer) String() string {
	return fmt.Sprintf("<%s host=%s port=%d auth=%s >", s.Type(), s.Host, s.Port, s.Auth)
}

func (s *HttpServer) Type() string { return "http" }

func (s *HttpServer) Prepare() error {
	return nil // No preparation needed
}

func (s *HttpServer) Prepared() bool { return true }

func (s *HttpServer) Connect(addr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(s.Host, strconv.Itoa(s.Port)))
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	req, err := http.NewRequest("CONNECT", "http://"+addr, nil)
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	if s.Auth != nil {
		req.Header.Add("proxy-authorization", "Basic "+s.Auth.Base64())
	}

	err = req.Write(conn)
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	res, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	if res.StatusCode != 200 {
		return nil, errtrace.Errorf("Cannot connect to %s\nStatus: %d - %s", addr, res.StatusCode, res.Status)
	}

	return conn, nil
}
