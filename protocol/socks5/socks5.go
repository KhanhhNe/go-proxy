package socks5

import (
	"bufio"
	"go-proxy/rwutil"
	"io"
	"net"

	"braces.dev/errtrace"
)

const (
	VER_SOCKS5 byte = 0x05

	AUTH_NoAuth             byte = 0x00
	AUTH_UsernamePassword   byte = 0x02
	AUTH_NoAcceptableMethod byte = 0xFF

	AUTH_VER_UsernamePassword byte = 0x01

	ADDR_IPv4       byte = 0x01
	ADDR_DomainName byte = 0x03
	ADDR_IPv6       byte = 0x04

	CMD_Connect      byte = 0x01
	CMD_Bind         byte = 0x02
	CMD_UdpAssociate byte = 0x03

	REP_Succeeded            byte = 0x00 // X'00' succeeded
	REP_GeneralFailure       byte = 0x01 // X'01' general SOCKS server failure
	REP_ConnectionNotAllowed byte = 0x02 // X'02' connection not allowed by ruleset
	REP_NetworkUnreachable   byte = 0x03 // X'03' Network unreachable
	REP_HostUnreachable      byte = 0x04 // X'04' Host unreachable
	REP_ConnectionRefused    byte = 0x05 // X'05' Connection refused
	REP_TtlExpired           byte = 0x06 // X'06' TTL expired
	REP_CommandNotSupported  byte = 0x07 // X'07' Command not supported
	REP_AddrTypeNotSupported byte = 0x08 // X'08' Address type not supported
)

// +----+----------+----------+
// |VER | NMETHODS | METHODS  |
// +----+----------+----------+
// | 1  |    1     | 1 to 255 |
// +----+----------+----------+
type MSG_ClientConnect struct {
	Version  byte
	NMethods byte
	Methods  []byte
}

func Read_ClientConnect(r io.Reader) (MSG_ClientConnect, error) {
	msg := MSG_ClientConnect{}

	err := rwutil.Scan(r, &msg.Version, &msg.NMethods)
	if err != nil {
		return msg, errtrace.Wrap(err)
	}

	buf, err := rwutil.ScanBuf(r, int(msg.NMethods))
	if err != nil {
		return msg, errtrace.Wrap(err)
	}

	msg.Methods = buf
	return msg, nil
}

func Write_ClientConnect(w *bufio.Writer, m MSG_ClientConnect) error {
	return errtrace.Wrap(rwutil.WriteBytesFlush(w,
		[]byte{
			m.Version,
			m.NMethods,
		},
		m.Methods,
	))
}

// +----+--------+
// |VER | METHOD |
// +----+--------+
// | 1  |   1    |
// +----+--------+
type MSG_SelectMethod struct {
	Version byte
	Method  byte
}

func Read_SelectMethod(r io.Reader) (m MSG_SelectMethod, err error) {
	err = rwutil.Scan(r, &m.Version, &m.Method)
	return m, errtrace.Wrap(err)
}

func Write_SelectMethod(w *bufio.Writer, m MSG_SelectMethod) error {
	return errtrace.Wrap(
		rwutil.WriteBytesFlush(w, []byte{m.Version, m.Method}))
}

// +----+------+----------+------+----------+
// |VER | ULEN |  UNAME   | PLEN |  PASSWD  |
// +----+------+----------+------+----------+
// | 1  |  1   | 1 to 255 |  1   | 1 to 255 |
// +----+------+----------+------+----------+
type MSG_AuthUserPass struct {
	Version  byte
	UserLen  byte
	Username string
	PassLen  byte
	Password string
}

func Read_AuthUserPass(r io.Reader) (m MSG_AuthUserPass, err error) {
	err = rwutil.Scan(r, &m.Version, &m.UserLen)
	if err != nil {
		return m, errtrace.Wrap(err)
	}

	buf, err := rwutil.ScanBuf(r, int(m.UserLen))
	if err != nil {
		return m, errtrace.Wrap(err)
	}
	m.Username = string(buf)

	err = rwutil.Scan(r, &m.PassLen)
	if err != nil {
		return m, errtrace.Wrap(err)
	}

	buf, err = rwutil.ScanBuf(r, int(m.PassLen))
	if err != nil {
		return m, errtrace.Wrap(err)
	}

	m.Password = string(buf)

	return m, nil
}

func Write_AuthUserPass(w *bufio.Writer, m MSG_AuthUserPass) error {
	return errtrace.Wrap(rwutil.WriteBytesFlush(w,
		[]byte{
			m.Version,
			m.UserLen,
		},
		[]byte(m.Username),
		[]byte{
			m.PassLen,
		},
		[]byte(m.Password),
	))
}

// +----+--------+
// |VER | STATUS |
// +----+--------+
// | 1  |   1    |
// +----+--------+
type MSG_AuthUserPassReply struct {
	Version byte
	Status  byte
}

func Read_AuthUserPassReply(r io.Reader) (m MSG_AuthUserPassReply, err error) {
	err = rwutil.Scan(r, &m.Version, &m.Status)
	return m, errtrace.Wrap(err)
}

func Write_AuthUserPassReply(w *bufio.Writer, m MSG_AuthUserPassReply) error {
	return errtrace.Wrap(rwutil.WriteBytesFlush(w, []byte{
		m.Version,
		m.Status,
	}))
}

// +----+-----+-------+------+----------+----------+
// |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
// +----+-----+-------+------+----------+----------+
// | 1  |  1  | X'00' |  1   | Variable |    2     |
// +----+-----+-------+------+----------+----------+
type MSG_Command struct {
	Version  byte
	Command  byte
	Reserved byte
	AddrType byte
	DstAddr  string
	DstPort  uint16
}

func Read_Command(r io.Reader) (MSG_Command, error) {
	msg := MSG_Command{}

	err := rwutil.Scan(r, &msg.Version, &msg.Command, &msg.Reserved, &msg.AddrType)
	if err != nil {
		return msg, errtrace.Wrap(err)
	}

	switch msg.AddrType {
	case ADDR_IPv4:
		buf, err := rwutil.ScanBuf(r, 4)
		if err != nil {
			return msg, errtrace.Wrap(err)
		}

		msg.DstAddr = net.IP(buf).String()
	case ADDR_IPv6:
		buf, err := rwutil.ScanBuf(r, 16)
		if err != nil {
			return msg, errtrace.Wrap(err)
		}

		msg.DstAddr = net.IP(buf).String()
	case ADDR_DomainName:
		var n byte
		err := rwutil.Scan(r, &n)
		if err != nil {
			return msg, errtrace.Wrap(err)
		}

		buf, err := rwutil.ScanBuf(r, int(n))
		if err != nil {
			return msg, errtrace.Wrap(err)
		}

		msg.DstAddr = string(buf)
	}

	buf, err := rwutil.ScanBuf(r, 2)
	if err != nil {
		return msg, errtrace.Wrap(err)
	}

	msg.DstPort = uint16(buf[0])<<8 | uint16(buf[1])

	return msg, nil
}

func Write_Command(w *bufio.Writer, m MSG_Command) error {
	var addr []byte

	if m.AddrType == ADDR_DomainName {
		addr = append([]byte{byte(len(m.DstAddr))}, []byte(m.DstAddr)...)
	} else {
		addr = net.ParseIP(m.DstAddr)
		if m.AddrType == ADDR_IPv4 {
			addr = addr[12:]
		}
	}

	return errtrace.Wrap(rwutil.WriteBytesFlush(w,
		[]byte{
			m.Version,
			m.Command,
			m.Reserved,
			m.AddrType,
		},
		addr,
		[]byte{
			byte(m.DstPort >> 8),
			byte(m.DstPort),
		},
	))
}

// +----+-----+-------+------+----------+----------+
// |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
// +----+-----+-------+------+----------+----------+
// | 1  |  1  | X'00' |  1   | Variable |    2     |
// +----+-----+-------+------+----------+----------+
type MSG_CommandReply struct {
	Version  byte
	Reply    byte
	Reserved byte
	AddrType byte
	BindAddr string
	BindPort uint16
}

func Read_CommandReply(r io.Reader) (m MSG_CommandReply, err error) {
	err = rwutil.Scan(r, &m.Version, &m.Reply, &m.Reserved, &m.AddrType)
	if err != nil {
		return m, errtrace.Wrap(err)
	}

	var len byte
	switch m.AddrType {
	case ADDR_IPv4:
		len = 4
	case ADDR_IPv6:
		len = 6
	case ADDR_DomainName:
		err = rwutil.Scan(r, &len)
		if err != nil {
			return m, errtrace.Wrap(err)
		}
	}

	buf, err := rwutil.ScanBuf(r, int(len))
	if err != nil {
		return m, errtrace.Wrap(err)
	}

	m.BindAddr = string(buf)

	buf, err = rwutil.ScanBuf(r, 2)
	if err != nil {
		return m, errtrace.Wrap(err)
	}

	m.BindPort = uint16(buf[0])<<8 | uint16(buf[1])

	return m, nil
}

func Write_CommandReply(w *bufio.Writer, m MSG_CommandReply) error {
	var addr net.IP

	switch m.AddrType {
	case ADDR_IPv4, ADDR_IPv6:
		addr = net.ParseIP(m.BindAddr)
		if m.AddrType == ADDR_IPv4 {
			addr = addr[12:]
		}
	case ADDR_DomainName:
		addr = net.IP(m.BindAddr)
		addr = append([]byte{byte(len(addr))}, addr...)
	}

	var portBuf [2]byte
	portBuf[0] = byte(m.BindPort >> 8)
	portBuf[1] = byte(m.BindPort)

	return errtrace.Wrap(rwutil.WriteBytesFlush(w,
		[]byte{m.Version, m.Reply, m.Reserved, m.AddrType},
		addr,
		portBuf[:]))
}
