package proxyserver

import (
	"bufio"
	"go-proxy/protocol/socks5"
	"net"
	"strconv"

	"braces.dev/errtrace"
)

type ServerSocks5State struct {
}

func (s *Server) prepareSocks5() error {
	return nil // No preparation needed
}

func (s *Server) isPreparedSocks5() bool { return true }

func (s *Server) connectSocks5(target string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(target)
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

	conn, err := s.connectAndAuthSocks5()
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	success := false
	defer func() {
		if !success {
			conn.Close()
		}
	}()

	err = socks5.Write_Command(writer, socks5.MSG_Command{
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

	msg, err := socks5.Read_CommandReply(reader)
	if err != nil {
		return nil, errtrace.Wrap(err)
	}
	if msg.Reply != socks5.REP_Succeeded {
		return nil, errtrace.Errorf("Socks5 connect target failed. Status code: %x", msg.Reply)
	}

	success = true
	return conn, nil
}

func (s *Server) connectAndAuthSocks5() (net.Conn, error) {
	netConn, err := net.Dial("tcp", net.JoinHostPort(s.Host, strconv.Itoa(s.Port)))
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	reader := bufio.NewReader(netConn)
	writer := bufio.NewWriter(netConn)

	auth := socks5.AUTH_NoAuth
	if s.Auth != nil {
		auth = socks5.AUTH_UsernamePassword
	}

	err = socks5.Write_ClientConnect(writer, socks5.MSG_ClientConnect{
		Version:  socks5.VER_SOCKS5,
		NMethods: 1,
		Methods:  []byte{auth},
	})
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	msg, err := socks5.Read_SelectMethod(reader)
	if err != nil {
		return nil, errtrace.Wrap(err)
	}
	if msg.Method == socks5.AUTH_NoAcceptableMethod {
		return nil, errtrace.Errorf("Socks5 returned no acceptable methods")
	}

	if s.Auth != nil {
		err = socks5.Write_AuthUserPass(writer, socks5.MSG_AuthUserPass{
			Version:  socks5.AUTH_VER_UsernamePassword,
			UserLen:  byte(len(s.Auth.Username)),
			Username: s.Auth.Username,
			PassLen:  byte(len(s.Auth.Password)),
			Password: s.Auth.Password,
		})
		if err != nil {
			return nil, errtrace.Wrap(err)
		}

		msg, err := socks5.Read_AuthUserPassReply(reader)
		if err != nil {
			return nil, errtrace.Wrap(err)
		}

		if msg.Status != 0x00 {
			return nil, errtrace.Errorf("Socks5 authentication failed. Status code: %x", msg.Status)
		}
	}

	return netConn, nil
}
