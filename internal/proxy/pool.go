package proxy

import (
	"sync"

	"github.com/ODudek/ConnX/internal/loadbalancer"
)

type ServerPool struct {
    backends []*Backend
    balancer loadbalancer.Balancer
    mutex    sync.RWMutex
}

func NewServerPool(balancer loadbalancer.Balancer) *ServerPool {
    return &ServerPool{
        backends: make([]*Backend, 0),
        balancer: balancer,
    }
}

func (s *ServerPool) AddBackend(backend *Backend) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    s.backends = append(s.backends, backend)
}

func (s *ServerPool) GetNextPeer() *Backend {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    // Convert []*Backend to []loadbalancer.Backend
    lbBackends := make([]loadbalancer.Backend, len(s.backends))
    for i, b := range s.backends {
        lbBackends[i] = b
    }

    lbBackend := s.balancer.Next(lbBackends)
    if lbBackend == nil {
        return nil
    }

    // Convert back to *Backend
    backend := lbBackend.(*Backend)
    backend.IncrementRequests()
    backend.IncrementActiveConns()

    return backend
}

func (s *ServerPool) ReleaseBackend(backend *Backend) {
    if backend != nil {
        backend.DecrementActiveConns()
    }
}

func (s *ServerPool) MarkBackendStatus(backend *Backend, alive bool) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    backend.Alive = alive
}

func (s *ServerPool) GetBackends() []*Backend {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    return s.backends
}

func (s *ServerPool) SetBalancer(balancer loadbalancer.Balancer) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    s.balancer = balancer
}
