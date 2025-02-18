package proxy

import (
	"net/url"
	"sync/atomic"
)


type Backend struct {
    URL      *url.URL
    Alive    bool
    Requests uint64
    Errors   uint64
}

func NewBackend(urlStr string) (*Backend, error) {
    u, err := url.Parse(urlStr)
    if err != nil {
        return nil, err
    }

    return &Backend{
        URL:   u,
        Alive: true,
    }, nil
}

func (b *Backend) IncrementRequests() {
    atomic.AddUint64(&b.Requests, 1)
}

func (b *Backend) IncrementErrors() {
    atomic.AddUint64(&b.Errors, 1)
}

func (b *Backend) GetStats() (requests, errors uint64) {
    return atomic.LoadUint64(&b.Requests), atomic.LoadUint64(&b.Errors)
}
