package loadbalancer

import "sync/atomic"

type RoundRobin struct {
    current uint64
}

func (r *RoundRobin) Next(count int) int {
    return int(atomic.AddUint64(&r.current, 1) % uint64(count))
}
