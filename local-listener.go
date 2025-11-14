package main

import (
	"bufio"
	"context"
	"fmt"
	"go-proxy/common"
	"go-proxy/protocol/socks5"
	"go-proxy/rwutil"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"braces.dev/errtrace"
	psutilnet "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

type LocalListener struct {
	Port     int
	Listener net.Listener
	Auth     *common.ProxyAuth
	Filter   ServerFilter
	Stat     ListenerStat
}

type ListenerStat struct {
	Sent     uint64
	Received uint64
}

type IncomingConnection struct {
	*net.TCPConn

	Listener *LocalListener
	Process  *process.Process
}

func (c *IncomingConnection) Write(b []byte) (int, error) {
	n, err := c.TCPConn.Write(b)
	c.Listener.Stat.Sent += uint64(n)
	return n, err
}

func (c *IncomingConnection) Read(b []byte) (int, error) {
	n, err := c.TCPConn.Read(b)
	c.Listener.Stat.Received += uint64(n)
	return n, err
}

type DoneCallback func(err error)

func NewLocalListener(port int, auth *common.ProxyAuth, filter ServerFilter) *LocalListener {
	return &LocalListener{
		port,
		nil,
		auth,
		filter,
		ListenerStat{},
	}
}

func (l *LocalListener) Printlnf(f string, a ...any) {
	f = fmt.Sprintf("[LocalListener :%d] ", l.Port) + f + "\n"
	fmt.Printf(f, a...)
}

func FindTcpProcess(addr string) (*process.Process, error) {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	conns, err := psutilnet.Connections("tcp")
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	var pid int32
	for _, conn := range conns {
		if int(conn.Laddr.Port) == portInt {
			pid = conn.Pid
		}
	}
	if pid == 0 {
		return nil, nil
	}

	proc, err := process.NewProcess(pid)
	return proc, errtrace.Wrap(err)
}

func (l *LocalListener) Serve(ctx context.Context, cb DoneCallback) {
	config := net.ListenConfig{}
	listener, err := config.Listen(ctx, "tcp", net.JoinHostPort("0.0.0.0", strconv.Itoa(l.Port)))
	if err != nil {
		cb(errtrace.Wrap(err))
		return
	}
	l.Listener = listener

	l.Printlnf("Listening at %d", l.Port)

	running := true

	for running {
		c, err := l.Listener.Accept()
		if err != nil {
			cb(errtrace.Wrap(err))
		}
		netConn := c.(*net.TCPConn)

		go func() {
			defer netConn.Close()

			addr := netConn.RemoteAddr().String()
			proc, err := FindTcpProcess(addr)

			if err != nil || proc == nil {
				l.Printlnf("Cannot find associated TCP socks for port %s", addr)
				if err != nil {
					l.Printlnf("Error: %+v", err)
				}
			} else {
				name, err := proc.Name()
				if err != nil {
					l.Printlnf("Error: %+v", err)
				}
				l.Printlnf("Found process: %d %s for %s", proc.Pid, name, addr)
			}

			conn := &IncomingConnection{
				netConn,
				l,
				proc,
			}

			reader := bufio.NewReader(conn)
			writer := bufio.NewWriter(conn)

			version, err := reader.Peek(1)
			if err != nil {
				l.Printlnf("Proxy handler error: %+v", err)
			}

			switch version[0] {
			case socks5.VER_SOCKS5:
				err = l.HandleSocks5(conn, reader, writer)
			default:
				// If not recognized, it could be HTTP request
				err = l.HandleHttp(conn, reader, writer)
			}

			if err != nil {
				l.Printlnf("Proxy handler error: %+v", err)
			}
		}()
	}
}

func (l *LocalListener) HandleHttp(conn *IncomingConnection, reader *bufio.Reader, writer *bufio.Writer) error {
	req, err := http.ReadRequest(reader)
	if err != nil {
		return errtrace.Wrap(err)
	}

	res := http.Response{
		Proto:      req.Proto,
		ProtoMajor: req.ProtoMajor,
		ProtoMinor: req.ProtoMinor,
	}

	target := ""

	if req.Method == "CONNECT" {
		target = req.RequestURI
	} else {
		uri, err := url.Parse(req.RequestURI)
		if err != nil {
			// Wrong URL format passed - result in 400 Bad Request
			res.StatusCode = http.StatusBadRequest
			err := rwutil.WriteResponseFlush(writer, res)
			return errtrace.Wrap(err)
		}

		if !uri.IsAbs() {
			// Non-absolute request URI passed
			// (expect absolute URI for proxy server to proceed)
			// - result in 400 Bad Request
			res.StatusCode = http.StatusBadRequest
			err := rwutil.WriteResponseFlush(writer, res)
			return errtrace.Wrap(err)
		}

		target = uri.Host
		if uri.Port() == "" {
			target += ":80"
		}

		req.RequestURI = uri.RequestURI()
		if uri.Fragment != "" {
			req.RequestURI += "#" + uri.Fragment
		}
	}

	if l.Auth != nil {
		auth := req.Header.Get("proxy-authorization")

		if auth == "" || !l.Auth.VerifyBasic(auth) {
			res.StatusCode = http.StatusProxyAuthRequired
			res.Header.Add("proxy-authenticate", "Basic realm=\"GoProxy\"")
			err := rwutil.WriteResponseFlush(writer, res)
			if err != nil {
				return errtrace.Wrap(err)
			}
			return nil
		}

		req.Header.Del("proxy-authorization")
	}

	s, err := ListenerServerManager.GetServer(l.Filter)
	if err != nil {
		return errtrace.Wrap(err)
	}
	if !s.Server.IsPrepared() {
		err = s.Server.Prepare()
		if err != nil {
			return errtrace.Wrap(err)
		}
	}
	remoteConn, err := s.Server.Connect(target)
	if err != nil {
		res.StatusCode = http.StatusBadGateway
		er := rwutil.WriteResponseFlush(writer, res)
		if er != nil {
			return errtrace.Wrap(er)
		}

		return errtrace.Wrap(err)
	}

	defer remoteConn.Close()

	if req.Method == "CONNECT" {
		res.StatusCode = http.StatusOK
		res.Status = "Connection Etablished"
		err := rwutil.WriteResponseFlush(writer, res)
		if err != nil {
			return errtrace.Wrap(err)
		}
	} else {
		err = req.Write(remoteConn)
		if err != nil {
			return errtrace.Wrap(err)
		}
	}

	rwutil.TunnelConns(conn, remoteConn)
	return nil
}

func (l *LocalListener) HandleSocks5(conn *IncomingConnection, reader *bufio.Reader, writer *bufio.Writer) error {
	msg, err := socks5.Read_ClientConnect(reader)
	if err != nil {
		return errtrace.Wrap(err)
	}

	authType := socks5.AUTH_NoAuth
	if l.Auth != nil {
		authType = socks5.AUTH_UsernamePassword
	}

	methodMatched := false
	for _, m := range msg.Methods {
		if m == authType {
			methodMatched = true
		}
	}

	if !methodMatched {
		return errtrace.Wrap(socks5.Write_SelectMethod(writer, socks5.MSG_SelectMethod{
			Version: socks5.VER_SOCKS5,
			Method:  socks5.AUTH_NoAcceptableMethod,
		}))
	}

	err = socks5.Write_SelectMethod(writer, socks5.MSG_SelectMethod{
		Version: socks5.VER_SOCKS5,
		Method:  authType,
	})
	if err != nil {
		return errtrace.Wrap(err)
	}

	switch authType {
	case socks5.AUTH_NoAuth:
	case socks5.AUTH_UsernamePassword:
		msg, err := socks5.Read_AuthUserPass(reader)
		if err != nil {
			return errtrace.Wrap(err)
		}

		if msg.Username != l.Auth.Username || msg.Password != l.Auth.Password {
			// Non-zero means failure
			return socks5.Write_AuthUserPassReply(writer, socks5.MSG_AuthUserPassReply{
				Version: socks5.AUTH_VER_UsernamePassword,
				Status:  0xFF,
			})
		}

		// 0x00 indicates authentication succeeded
		err = socks5.Write_AuthUserPassReply(writer, socks5.MSG_AuthUserPassReply{
			Version: socks5.AUTH_VER_UsernamePassword,
			Status:  0x00,
		})
		if err != nil {
			return errtrace.Wrap(err)
		}
	}

	for {
		msg, err := socks5.Read_Command(reader)
		if err != nil {
			return errtrace.Wrap(err)
		}

		switch msg.Command {
		case socks5.CMD_Connect:
			s, err := ListenerServerManager.GetServer(l.Filter)
			if err != nil {
				return errtrace.Wrap(err)
			}
			if !s.Server.IsPrepared() {
				err = s.Server.Prepare()
				if err != nil {
					return errtrace.Wrap(err)
				}
			}
			remoteConn, err := s.Server.Connect(net.JoinHostPort(msg.DstAddr, strconv.Itoa(int(msg.DstPort))))

			if err != nil {
				er := socks5.Write_CommandReply(writer, socks5.MSG_CommandReply{
					Version:  socks5.VER_SOCKS5,
					Reply:    socks5.REP_GeneralFailure,
					AddrType: socks5.ADDR_IPv4,
					BindAddr: "127.0.0.1",
					BindPort: 0,
				})
				if er != nil {
					return errtrace.Wrap(er)
				}
				return errtrace.Wrap(err)
			}

			defer remoteConn.Close()

			addr := remoteConn.LocalAddr().(*net.TCPAddr)
			host, port, err := net.SplitHostPort(addr.String())
			if err != nil {
				return errtrace.Wrap(err)
			}

			portInt, err := strconv.Atoi(port)
			if err != nil {
				return errtrace.Wrap(err)
			}
			addrType := socks5.ADDR_IPv4
			if addr.IP.To4() == nil {
				addrType = socks5.ADDR_IPv6
			}

			err = socks5.Write_CommandReply(writer, socks5.MSG_CommandReply{
				Version:  socks5.VER_SOCKS5,
				Reply:    socks5.REP_Succeeded,
				AddrType: addrType,
				BindAddr: host,
				BindPort: uint16(portInt),
			})
			if err != nil {
				return errtrace.Wrap(err)
			}

			rwutil.TunnelConns(conn, remoteConn)
			return nil
		default:
			return errtrace.Wrap(socks5.Write_CommandReply(writer, socks5.MSG_CommandReply{
				Version:  socks5.VER_SOCKS5,
				Reply:    socks5.REP_CommandNotSupported,
				AddrType: socks5.ADDR_IPv4,
				BindAddr: "127.0.0.1",
				BindPort: 0,
			}))
		}
	}
}
