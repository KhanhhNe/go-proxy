package main

import (
	"context"
	"fmt"
	"go-proxy/common"
	"go-proxy/proxyserver"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type MyService struct {
	app *application.App
}

func NewMyService(app *application.App) *MyService {
	return &MyService{app: app}
}

func (s *MyService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
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

	var wg sync.WaitGroup
	wg.Go(func() {
		ListenerServerManager.Serve()
	})
	<-time.After(time.Second)

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
		proxyserver.NewServer("64.137.75.244", 6164, &common.ProxyAuth{
			Username: "fuaultwx",
			Password: "frqbhyi7fs1a",
		}),
	})

	for _, s := range ListenerServerManager.Servers {
		fmt.Println(s)
	}

	return nil
}

func (s *MyService) GetManager() *listenerServerManager {
	return ListenerServerManager
}
