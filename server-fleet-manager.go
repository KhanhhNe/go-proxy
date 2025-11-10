package main

import (
	"context"
	"go-proxy/proxyserver"
	"iter"
	"maps"
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
	Server proxyserver.Server

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
	proxyserver.NewDirectProxyServer(),
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

func (m *listenerServerManager) AddServers(servers []proxyserver.Server) {
	for _, s := range servers {
		managedServer := &ManagedProxyServer{
			s,
			map[string]bool{},
		}
		managedServer.AddTags(s.Type())

		m.Servers[s.String()] = managedServer
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

		if len(filter.Tags) > 0 {
			found := false
			for _, t := range filter.Tags {
				v, ok := s.Tags[t]
				if ok && v {
					found = true
					break
				}
			}
			if !found {
				continue
			}
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

func (m *listenerServerManager) Serve() {
	m.IsServing = true
	m.serveInactiveListeners()
	m.Wg.Wait()
}
