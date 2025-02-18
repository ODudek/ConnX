package proxy

import (
	"sync"
	"sync/atomic"
)

type ServerPool struct {
    backends []*Backend
    current  uint64
    mutex    sync.RWMutex
}

func NewServerPool() *ServerPool {
    return &ServerPool{
        backends: make([]*Backend, 0),
    }
}

func (s *ServerPool) AddBackend(backend *Backend) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    s.backends = append(s.backends, backend)
}

func (s *ServerPool) NextIndex() int {
    return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

func (s *ServerPool) GetNextPeer() *Backend {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    next := s.NextIndex()
    l := len(s.backends)
    for i := 0; i < l; i++ {
        idx := (next + i) % l
        if s.backends[idx].Alive {
            s.backends[idx].IncrementRequests()
            return s.backends[idx]
        }
    }
    return nil
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
