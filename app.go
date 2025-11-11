package main

import (
	"context"
	"fmt"
	"go-proxy/common"
	"go-proxy/proxyserver"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	ListenerServerManager.AddServers([]*proxyserver.Server{
		proxyserver.NewServer("localhost", 2222, &common.ProxyAuth{
			Username: "ubuntu",
			Password: "ubuntu",
		}),
		proxyserver.NewServer("127.0.0.1", 9001, &common.ProxyAuth{
			Username: "khanh",
			Password: "khanh",
		}),
		proxyserver.NewServer("::1", 9002, &common.ProxyAuth{
			Username: "khanh",
			Password: "khanh",
		}),
	})

	for _, s := range ListenerServerManager.Servers {
		fmt.Println(s)
		switch s.Server.Port {
		case 2222:
			s.AddTags("ssh")
			s.Server.Protocols[proxyserver.PROTO_Ssh] = true
		case 9001:
			s.AddTags("socks5")
			s.Server.Protocols[proxyserver.PROTO_Socks5] = true
		case 9002:
			s.AddTags("http")
			s.Server.Protocols[proxyserver.PROTO_Http] = true
		}
	}

	ListenerServerManager.AddListeners([]*LocalListener{
		NewLocalListener(8000, &common.ProxyAuth{
			Username: "khanh",
			Password: "khanh",
		}, ServerFilter{Tags: []string{"ssh"}}),
		NewLocalListener(8001, &common.ProxyAuth{
			Username: "khanh",
			Password: "khanh",
		}, ServerFilter{Tags: []string{"http"}}),
		NewLocalListener(8002, &common.ProxyAuth{
			Username: "khanh",
			Password: "khanh",
		}, ServerFilter{Tags: []string{"socks5"}}),

		NewLocalListener(9001, &common.ProxyAuth{
			Username: "khanh",
			Password: "khanh",
		}, ServerFilter{IgnoreAll: true}),
		NewLocalListener(9002, &common.ProxyAuth{
			Username: "khanh",
			Password: "khanh",
		}, ServerFilter{IgnoreAll: true}),
	})

	ListenerServerManager.Serve()
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
