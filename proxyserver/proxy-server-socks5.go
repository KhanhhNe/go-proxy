package proxyserver

import (
	"bufio"
	"fmt"
	"go-proxy/common"
	"go-proxy/protocol/socks5"
	"net"
	"strconv"
	"time"

	"braces.dev/errtrace"
)

type Socks5Server struct {
	Host    string
	Port    int
	Timeout time.Duration
	Auth    *common.ProxyAuth
}

type Socks5Conn struct {
	net.Conn

	Reader *bufio.Reader
	Writer *bufio.Writer
}

// Enforce type check
var _ Server = &Socks5Server{}

func NewSocks5ProxyServer(host string, port int, auth *common.ProxyAuth) *Socks5Server {
	return &Socks5Server{
		Host:    host,
		Port:    port,
		Auth:    auth,
		Timeout: 30 * time.Second,
	}
}

func (s *Socks5Server) String() string {
	return fmt.Sprintf("<%s host=%s port=%d auth=%s >", s.Type(), s.Host, s.Port, s.Auth)
}

func (s *Socks5Server) Type() string { return "socks5" }

func (s *Socks5Server) Prepare() error {
	return nil // No preparation needed
}

func (s *Socks5Server) Prepared() bool { return true }

func (s *Socks5Server) Connect(addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	var addrType byte
	ip := net.ParseIP(host)
	if ip != nil {
		if net.IP(ip).To4() != nil {
			addrType = socks5.ADDR_IPv4
		} else {
			addrType = socks5.ADDR_IPv6
		}
	} else {
		addrType = socks5.ADDR_DomainName
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	conn, err := s.ConnectAndAuth()
	success := false
	defer func() {
		if !success {
			conn.Close()
		}
	}()

	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	err = socks5.Write_Command(conn.Writer, socks5.MSG_Command{
		Version:  socks5.VER_SOCKS5,
		Command:  socks5.CMD_Connect,
		Reserved: 0x00,
		AddrType: addrType,
		DstAddr:  host,
		DstPort:  uint16(portNum),
	})
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	msg, err := socks5.Read_CommandReply(conn.Reader)
	if err != nil {
		return nil, errtrace.Wrap(err)
	}
	if msg.Reply != socks5.REP_Succeeded {
		return nil, errtrace.Errorf("Socks5 connect target failed. Status code: %x", msg.Reply)
	}

	success = true
	return conn, nil
}

func (s *Socks5Server) ConnectAndAuth() (*Socks5Conn, error) {
	netConn, err := net.Dial("tcp", net.JoinHostPort(s.Host, strconv.Itoa(s.Port)))
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	conn := Socks5Conn{
		Conn:   netConn,
		Reader: bufio.NewReader(netConn),
		Writer: bufio.NewWriter(netConn),
	}

	auth := socks5.AUTH_NoAuth
	if s.Auth != nil {
		auth = socks5.AUTH_UsernamePassword
	}

	err = socks5.Write_ClientConnect(conn.Writer, socks5.MSG_ClientConnect{
		Version:  socks5.VER_SOCKS5,
		NMethods: 1,
		Methods:  []byte{auth},
	})
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	msg, err := socks5.Read_SelectMethod(conn.Reader)
	if err != nil {
		return nil, errtrace.Wrap(err)
	}
	if msg.Method == socks5.AUTH_NoAcceptableMethod {
		return nil, errtrace.Errorf("Socks5 returned no acceptable methods")
	}

	if s.Auth != nil {
		err = socks5.Write_AuthUserPass(conn.Writer, socks5.MSG_AuthUserPass{
			Version:  socks5.AUTH_VER_UsernamePassword,
			UserLen:  byte(len(s.Auth.Username)),
			Username: s.Auth.Username,
			PassLen:  byte(len(s.Auth.Password)),
			Password: s.Auth.Password,
		})
		if err != nil {
			return nil, errtrace.Wrap(err)
		}

		msg, err := socks5.Read_AuthUserPassReply(conn.Reader)
		if err != nil {
			return nil, errtrace.Wrap(err)
		}

		if msg.Status != 0x00 {
			return nil, errtrace.Errorf("Socks5 authentication failed. Status code: %x", msg.Status)
		}
	}

	return &conn, nil
}
