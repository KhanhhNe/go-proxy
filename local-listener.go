package main

import (
	"bufio"
	"context"
	"fmt"
	"go-proxy/common"
	"go-proxy/protocol/prx_http"
	"go-proxy/protocol/socks5"
	"go-proxy/rwutil"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"braces.dev/errtrace"
	"github.com/cakturk/go-netstat/netstat"
)

type LocalListener struct {
	Port     int
	Listener net.Listener
	Auth     *common.ProxyAuth
	Filter   ServerFilter
}

type IncomingConnection struct {
	net.Conn

	Process *netstat.Process
}

type DoneCallback func(err error)

func NewLocalListener(port int, auth *common.ProxyAuth, filter ServerFilter) *LocalListener {
	return &LocalListener{
		port,
		nil,
		auth,
		filter,
	}
}

func (l *LocalListener) Printlnf(f string, a ...any) {
	f = fmt.Sprintf("[LocalListener :%d] ", l.Port) + f + "\n"
	fmt.Printf(f, a...)
}

func FindTcpProcess(addr string) (*netstat.Process, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	ip := net.ParseIP(host)
	accept := func(e *netstat.SockTabEntry) bool {
		return strconv.Itoa(int(e.LocalAddr.Port)) == port
	}

	var procList []netstat.SockTabEntry
	if ip.To4() != nil {
		procList, err = netstat.TCPSocks(accept)
	} else {
		procList, err = netstat.TCP6Socks(accept)
	}
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	if len(procList) > 0 {
		return procList[0].Process, nil
	}

	return nil, nil
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
		netConn, err := l.Listener.Accept()
		if err != nil {
			cb(errtrace.Wrap(err))
		}

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
				l.Printlnf("Found process: %d %s for %s", proc.Pid, proc.Name, addr)
			}

			conn := &IncomingConnection{
				Conn:    netConn,
				Process: proc,
			}

			reader := bufio.NewReader(netConn)
			writer := bufio.NewWriter(netConn)

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

	target := ""

	if req.Method == "CONNECT" {
		target = req.RequestURI
	} else {
		uri, err := url.Parse(req.RequestURI)
		if err != nil {
			// Wrong URL format passed - result in 400 Bad Request
			err := rwutil.WriteStringFlush(
				writer,
				prx_http.MSG_BadRequest(),
			)
			return errtrace.Wrap(err)
		}

		if !uri.IsAbs() {
			// Non-absolute request URI passed
			// (expect absolute URI for proxy server to proceed)
			// - result in 400 Bad Request
			err := rwutil.WriteStringFlush(
				writer,
				prx_http.MSG_BadRequest(),
			)
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
		if auth == "" {
			err := rwutil.WriteStringFlush(
				writer,
				prx_http.MSG_ProxyAuthRequired(),
			)
			if err != nil {
				return errtrace.Wrap(err)
			}
			return nil
		}

		if !l.Auth.VerifyBasic(auth) {
			err := rwutil.WriteStringFlush(
				writer,
				prx_http.MSG_ProxyAuthRequired(),
			)
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
	err = s.Server.Prepare()
	if err != nil {
		return errtrace.Wrap(err)
	}
	remoteConn, err := s.Server.Connect(target)
	if err != nil {
		er := rwutil.WriteStringFlush(
			writer,
			prx_http.MSG_BadGateway(),
		)
		if er != nil {
			return errtrace.Wrap(er)
		}

		return errtrace.Wrap(err)
	}

	defer remoteConn.Close()

	if req.Method == "CONNECT" {
		err := rwutil.WriteStringFlush(
			writer,
			prx_http.MSG_ConnectionEtablished(),
		)
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
			err = s.Server.Prepare()
			if err != nil {
				return errtrace.Wrap(err)
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
