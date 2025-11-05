package proxy

import (
	"net/url"
	"sync/atomic"

	"github.com/ODudek/ConnX/internal/circuitbreaker"
)


type Backend struct {
    URL            *url.URL
    Alive          bool
    Requests       uint64
    Errors         uint64
    ActiveConns    int64  // For least connections algorithm
    Weight         int    // For weighted round-robin
    CircuitBreaker *circuitbreaker.CircuitBreaker
}

func NewBackend(urlStr string, weight int) (*Backend, error) {
    u, err := url.Parse(urlStr)
    if err != nil {
        return nil, err
    }

    if weight <= 0 {
        weight = 1
    }

    cb := circuitbreaker.NewCircuitBreaker(circuitbreaker.Config{
        MaxFailures:     5,
        Timeout:         60,
        ResetTimeout:    30,
        HalfOpenSuccess: 2,
    })

    return &Backend{
        URL:            u,
        Alive:          true,
        Weight:         weight,
        ActiveConns:    0,
        CircuitBreaker: cb,
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

func (b *Backend) IncrementActiveConns() {
    atomic.AddInt64(&b.ActiveConns, 1)
}

func (b *Backend) DecrementActiveConns() {
    atomic.AddInt64(&b.ActiveConns, -1)
}

func (b *Backend) GetActiveConns() int64 {
    return atomic.LoadInt64(&b.ActiveConns)
}

func (b *Backend) IsAvailable() bool {
    return b.Alive && b.CircuitBreaker.IsAllowed()
}

func (b *Backend) GetWeight() int {
    return b.Weight
}
