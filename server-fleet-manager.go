package main

import (
	"context"
	"go-proxy/common"
	"go-proxy/proxyserver"
	"go-proxy/threadpool"
	"iter"
	"maps"
	"net/netip"
	"sync"

	"braces.dev/errtrace"
)

type ManagedLocalListener struct {
	Listener *LocalListener

	ctx       context.Context
	cancel    context.CancelFunc
	IsServing bool
}

type ManagedProxyServer struct {
	Server *proxyserver.Server

	Tags map[string]bool
}

type listenerServerManager struct {
	Listeners map[int]*ManagedLocalListener
	Servers   map[string]*ManagedProxyServer

	IsServing bool
	Wg        sync.WaitGroup
}

type ServerFilter struct {
	Tags      []string
	IgnoreAll bool
}

var DirectProxy = &ManagedProxyServer{
	proxyserver.NewDirectServer(),
	map[string]bool{},
}

var ListenerServerManager = NewListenerServerManager()

func NewListenerServerManager() (s *listenerServerManager) {
	s = &listenerServerManager{
		map[int]*ManagedLocalListener{},
		map[string]*ManagedProxyServer{},
		false,
		sync.WaitGroup{},
	}
	return s
}

func (s *ManagedProxyServer) AddTags(tags ...string) {
	for _, t := range tags {
		s.Tags[t] = true
	}
}

func (s *ManagedProxyServer) HasAllTags(tags []string) bool {
	for _, t := range tags {
		hasTag, ok := s.Tags[t]
		if !ok || !hasTag {
			return false
		}
	}

	return true
}

func (m *listenerServerManager) AddServers(servers []*proxyserver.Server) {
	for _, s := range servers {
		managedServer := &ManagedProxyServer{
			s,
			map[string]bool{},
		}
		m.Servers[s.String()] = managedServer

		managedServer.checkServer()
	}
}

func (m *listenerServerManager) GetServer(filter ServerFilter) (*ManagedProxyServer, error) {
	if filter.IgnoreAll {
		return DirectProxy, nil
	}

	if len(m.Servers) == 0 {
		return nil, errtrace.Errorf("No more servers inside manager")
	}

	next, stop := iter.Pull(maps.Values(m.Servers))
	defer stop()

	for {
		s, valid := next()
		if !valid {
			break
		}

		if !s.HasAllTags(filter.Tags) {
			continue
		}

		return s, nil
	}

	return nil, errtrace.Errorf("Cannot get server")
}

func (m *listenerServerManager) AddListeners(listeners []*LocalListener) {
	for _, l := range listeners {
		ctx, cancel := context.WithCancel(context.Background())

		m.Listeners[l.Port] = &ManagedLocalListener{
			l,
			ctx,
			cancel,
			false,
		}
	}

	if m.IsServing {
		m.serveInactiveListeners()
	}
}

func (m *listenerServerManager) serveInactiveListeners() {
	for _, l := range m.Listeners {
		if l.IsServing {
			continue
		}

		m.Wg.Add(1)
		go func(ctx context.Context) {
			l.Listener.Serve(l.ctx, func(err error) {
				// Ignore error
				m.Wg.Done()
			})
		}(l.ctx)
	}
}

type CheckServerThread struct {
	server *ManagedProxyServer
}

func (t *CheckServerThread) Run() {
	s := t.server

	s.Server.CheckServer()
	for proto, supported := range s.Server.Protocols {
		if supported {
			s.AddTags(proto)
		}
	}

	if s.Server.PublicIp != "" {
		ip, err := netip.ParseAddr(s.Server.PublicIp)
		if err != nil {
			return
		}

		countryCode, err := common.GetIpCountry(ip)
		if err != nil {
			s.Server.Printlnf("Error getting IP country: IP %s, error: %+v", s.Server.PublicIp, err)
			return
		}

		s.AddTags(countryCode)
	}
}

var CheckServerPool = threadpool.NewThreadPool[*CheckServerThread](50)

func (s *ManagedProxyServer) checkServer() {
	CheckServerPool.AddTask(&CheckServerThread{s})
}

func (m *listenerServerManager) Serve() {
	m.IsServing = true
	m.serveInactiveListeners()
	m.Wg.Wait()
}
