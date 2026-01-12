package main

import (
	"context"
	"fmt"
	"go-proxy/common"
	"go-proxy/proxyserver"
	"net"
	"strconv"
	"strings"
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
	l1, _ := NewLocalListener(8000, &common.ProxyAuth{
		Username: "khanh",
		Password: "khanh",
	}, ServerFilter{Tags: []string{"ssh"}})
	l2, _ := NewLocalListener(8001, &common.ProxyAuth{
		Username: "khanh",
		Password: "khanh",
	}, ServerFilter{Tags: []string{"http"}})
	l3, _ := NewLocalListener(8002, &common.ProxyAuth{
		Username: "khanh",
		Password: "khanh",
	}, ServerFilter{Tags: []string{"socks5"}})
	l4, _ := NewLocalListener(9001, &common.ProxyAuth{
		Username: "khanh",
		Password: "khanh",
	}, ServerFilter{IgnoreAll: true})
	l5, _ := NewLocalListener(9002, &common.ProxyAuth{
		Username: "khanh",
		Password: "khanh",
	}, ServerFilter{IgnoreAll: true})

	ListenerServerManager.AddListeners([]*LocalListener{l1, l2, l3, l4, l5})

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
		proxyserver.NewServer("p.webshare.io", 80, &common.ProxyAuth{
			Username: "teuuumot12z-1",
			Password: "xjlhl0qpepassf",
		}),
	})

	for _, s := range ListenerServerManager.Servers {
		fmt.Println(s)
	}

	return nil
}

func (s *MyService) GetManager() *listenerServerManager {
	common.DataMutex.RLock()
	defer common.DataMutex.RUnlock()
	return ListenerServerManager
}

type AppState struct {
	LocalIp string
}

func (s *MyService) GetAppState() (state AppState) {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	wg.Go(func() {
		mu.Lock()
		state.LocalIp = getLocalIp()
		mu.Unlock()
	})

	wg.Wait()

	return
}

func (s *MyService) DeleteServers(ids []string) {
	for _, id := range ids {
		common.DataMutex.RLock()
		server, ok := ListenerServerManager.Servers[id]
		common.DataMutex.RUnlock()

		if ok {
			// Server shutdown
			server.Server.Cleanup()
		}

		common.DataMutex.Lock()
		delete(ListenerServerManager.Servers, id)
		common.DataMutex.Unlock()
	}
}

func (s *MyService) DeleteListeners(ports []int) {
	for _, port := range ports {
		common.DataMutex.RLock()
		listener, ok := ListenerServerManager.Listeners[port]
		common.DataMutex.RUnlock()

		if ok {
			// Stop listener
			listener.cancel()
		}

		common.DataMutex.Lock()
		delete(ListenerServerManager.Listeners, port)
		common.DataMutex.Unlock()
	}
}

func (s *MyService) ImportProxyFile(content, sep string, skipCol, defaultPort int, skipHeader bool) error {
	content = strings.TrimSpace(content)
	lines := strings.Split(content, "\n")
	if skipHeader {
		lines = lines[1:]
	}

	servers := make([]*proxyserver.Server, 0, len(lines))
	for _, line := range lines {
		s := s.ParseProxyLine(line, sep, skipCol, defaultPort)
		if s != nil {
			servers = append(servers, s)
		}
	}

	return ListenerServerManager.AddServers(servers)
}

func (s *MyService) ParseProxyLine(proxyStr, sep string, skip int, defaultPort int) *proxyserver.Server {
	parts := strings.Split(proxyStr, sep)
	parts = parts[skip:]

	host := ""
	port := 0
	var auth *common.ProxyAuth = nil

	for i, p := range parts {
		switch i {
		case 0:
			host = p
		case 1:
			prt, err := strconv.Atoi(p)
			if err == nil {
				port = prt
			}
		case 2:
			auth = &common.ProxyAuth{}
			auth.Username = p
		case 3:
			auth.Password = p
		}
	}

	if port == 0 {
		port = defaultPort
	}

	return proxyserver.NewServer(host, port, auth)
}

func (s *MyService) RecheckServer(id string) {
	ListenerServerManager.Servers[id].checkServer()
}

func getLocalIp() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "localhost"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
