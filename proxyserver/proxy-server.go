package proxyserver

import "net"

type Server interface {
	String() string
	Type() string

	// Prepare the server, authentication, etc
	Prepare() error
	Prepared() bool

	// Initiates connection to target server (usually web server)
	Connect(addr string) (net.Conn, error)
}
