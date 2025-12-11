package proxyserver

import (
	"bufio"
	"net"
	"net/http"
	"strconv"

	"braces.dev/errtrace"
)

type ServerHttpState struct {
}

func (s *Server) prepareHttp() error {
	return nil // No preparation needed
}

func (s *Server) isPreparedHttp() bool { return true }

func (s *Server) connectHttp(target string) (net.Conn, error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(s.Host, strconv.Itoa(s.Port)))
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	req, err := http.NewRequest("CONNECT", "http://"+target, nil)
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
		return nil, errtrace.Errorf("Cannot connect to %s\nStatus: %d - %s", target, res.StatusCode, res.Status)
	}

	return conn, nil
}

func (s *Server) cleanupHttp() {}
