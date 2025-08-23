package health

import (
	"net/http"
	"time"

	"github.com/ODudek/ConnX/internal/metrics"
	"github.com/ODudek/ConnX/internal/proxy"
)


type Checker struct {
    pool     *proxy.ServerPool
    interval time.Duration
    timeout  time.Duration
    metrics  *metrics.Metrics
}

func NewChecker(pool *proxy.ServerPool, interval, timeout int, metrics *metrics.Metrics) *Checker {
    return &Checker{
        pool:     pool,
        interval: time.Duration(interval) * time.Second,
        timeout:  time.Duration(timeout) * time.Second,
        metrics:  metrics,
    }
}

func (c *Checker) Start() {
    ticker := time.NewTicker(c.interval)
    go func() {
        for range ticker.C {
            c.CheckHealth()
        }
    }()
}

func (c *Checker) CheckHealth() {
    client := http.Client{
        Timeout: c.timeout,
    }

    for _, backend := range c.pool.GetBackends() {
        alive := c.isBackendAlive(&client, backend)
        c.pool.MarkBackendStatus(backend, alive)
        // Update metrics with backend status
        if c.metrics != nil {
            c.metrics.UpdateBackendStatus(backend.URL.String(), alive)
        }
    }
}

func (c *Checker) isBackendAlive(client *http.Client, backend *proxy.Backend) bool {
    resp, err := client.Get(backend.URL.String())
    if err != nil {
        backend.IncrementErrors()
        return false
    }
    defer resp.Body.Close()
    return resp.StatusCode == http.StatusOK
}
