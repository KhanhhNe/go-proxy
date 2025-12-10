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
	"time"

	"braces.dev/errtrace"
)

type ManagedLocalListener struct {
	Listener *LocalListener

	ctx    context.Context
	cancel context.CancelFunc
}

type ManagedProxyServer struct {
	Server *proxyserver.Server

	Tags map[string]bool
}

type listenerServerManager struct {
	Listeners map[int]*ManagedLocalListener
	Servers   map[string]*ManagedProxyServer

	ServerRecheckInterval time.Duration
	IsServing             bool
	Wg                    sync.WaitGroup
}

type ServerFilter struct {
	Tags      []string
	ServerIds map[string]bool
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
		60 * time.Second,
		false,
		sync.WaitGroup{},
	}
	return s
}

func (s *ManagedProxyServer) AddTags(tags ...string) {
	common.DataMutex.Lock()
	defer common.DataMutex.Unlock()

	for _, t := range tags {
		s.Tags[t] = true
	}
}

func (s *ManagedProxyServer) HasAllTags(tags []string) bool {
	common.DataMutex.RLock()
	defer common.DataMutex.RUnlock()

	for _, t := range tags {
		hasTag, ok := s.Tags[t]
		if !ok || !hasTag {
			return false
		}
	}

	return true
}

func (m *listenerServerManager) AddServers(servers []*proxyserver.Server) error {
	for _, s := range servers {
		managedServer := &ManagedProxyServer{
			s,
			map[string]bool{},
		}

		common.DataMutex.Lock()
		m.Servers[s.String()] = managedServer
		common.DataMutex.Unlock()

		managedServer.checkServer()

		listener, err := NewLocalListener(0, nil, ServerFilter{ServerIds: map[string]bool{s.Id: true}})
		if err != nil {
			return errtrace.Wrap(err)
		}
		m.AddListeners([]*LocalListener{listener})
	}

	return nil
}

func (m *listenerServerManager) GetServer(filter ServerFilter) (*ManagedProxyServer, error) {
	common.DataMutex.RLock()
	defer common.DataMutex.RUnlock()

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

		if len(filter.ServerIds) > 0 {
			if _, idAllowed := filter.ServerIds[s.Server.Id]; !idAllowed {
				continue
			}
		}

		return s, nil
	}

	return nil, errtrace.Errorf("Cannot get server")
}

func (m *listenerServerManager) AddListeners(listeners []*LocalListener) {
	common.DataMutex.Lock()

	for _, l := range listeners {
		ctx, cancel := context.WithCancel(context.Background())

		m.Listeners[l.Port] = &ManagedLocalListener{
			l,
			ctx,
			cancel,
		}
	}

	common.DataMutex.Unlock()

	if m.IsServing {
		m.serveInactiveListeners()
	}
}

func (m *listenerServerManager) serveInactiveListeners() {
	common.DataMutex.RLock()
	defer common.DataMutex.RUnlock()

	for _, l := range m.Listeners {
		if l.Listener.IsServing {
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

func (m *listenerServerManager) autoRecheckServers() {
	for {
		<-time.After(time.Second)
		since := time.Now().Add(-m.ServerRecheckInterval)

		common.DataMutex.RLock()

		for _, s := range m.Servers {
			if s.Server.LastChecked.After(since) {
				continue
			}

			s.checkServer()
		}

		common.DataMutex.RUnlock()
	}
}

type CheckServerThread struct {
	server *ManagedProxyServer
}

func (t *CheckServerThread) Id() string {
	return t.server.Server.String()
}

func (t *CheckServerThread) Run() {
	s := t.server

	s.Server.CheckServer()

	common.DataMutex.RLock()
	protos := s.Server.Protocols
	publicIp := s.Server.PublicIp
	common.DataMutex.RUnlock()

	for proto, supported := range protos {
		if supported {
			s.AddTags(proto)
		}
	}

	if publicIp != "" {
		ip, err := netip.ParseAddr(publicIp)
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
	common.DataMutex.Lock()
	m.IsServing = true
	common.DataMutex.Unlock()

	m.serveInactiveListeners()
	m.Wg.Go(m.autoRecheckServers)
	m.Wg.Wait()
}
