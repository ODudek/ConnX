# Benchmarking Guide

## Prerequisites

### Install Testing Tools

```bash
# macOS
brew install wrk
go install github.com/rakyll/hey@latest
go install github.com/tsenart/vegeta@latest

# Ubuntu/Debian
sudo apt install wrk apache2-utils
go install github.com/rakyll/hey@latest
go install github.com/tsenart/vegeta@latest
```

### System Tuning (Linux)

Before running high-load tests, tune your system:

```bash
# Increase file descriptors
ulimit -n 65536

# Increase local port range
sudo sysctl -w net.ipv4.ip_local_port_range="1024 65535"

# Increase max connections
sudo sysctl -w net.core.somaxconn=65536

# TCP tuning
sudo sysctl -w net.ipv4.tcp_tw_reuse=1
sudo sysctl -w net.ipv4.tcp_fin_timeout=15

# For high-load testing
sudo sysctl -w fs.file-max=2097152
```

## Quick Test

```bash
# 1. Start backends
python3 test/performance/simple_backend.py 8081 &
python3 test/performance/simple_backend.py 8082 &
python3 test/performance/simple_backend.py 8083 &

# 2. Start ConnX
./bin/proxy -config=test/performance/config.yaml

# 3. Run basic test
./test/performance/run_basic_test.sh
```

## Test Scenarios

### 1. Baseline Performance

Measure raw throughput with minimal load:

```bash
# Using wrk (recommended)
wrk -t4 -c100 -d30s --latency http://localhost:8080

# Using hey
hey -n 100000 -c 100 http://localhost:8080

# Using vegeta (constant rate)
echo "GET http://localhost:8080" | vegeta attack -rate=5000 -duration=30s | vegeta report
```

**Expected Results (4-core CPU, 16GB RAM):**
- RPS: 50,000-100,000+
- p50 latency: <2ms
- p99 latency: <10ms
- Memory: <100MB

### 2. Load Test (Find Breaking Point)

```bash
./test/performance/load_test.sh
```

This gradually increases load from 10 to 5,000 concurrent connections.

**What to look for:**
- Point where latency sharply increases
- Point where error rate increases
- Memory growth pattern

### 3. Stress Test

```bash
./test/performance/stress_test.sh
```

Pushes system beyond normal limits with 10,000 concurrent connections.

**What to monitor:**
- Error rate
- Connection failures
- System resource exhaustion
- Recovery after load drops

### 4. Algorithm Comparison

```bash
./test/performance/compare_algorithms.sh
```

Compares performance of:
- Round-robin
- Least connections
- Weighted round-robin

**Expected differences:**
- Round-robin: Fastest (simplest)
- Least-connections: ~5-10% slower (tracking overhead)
- Weighted-round-robin: ~3-7% slower (weight calculation)

### 5. Feature Impact Test

Test performance impact of each feature:

```bash
# Baseline (all features disabled)
# Edit config.yaml: disable circuit breaker, rate limit

# Test 1: Circuit breaker only
# Edit config.yaml: enable circuit breaker

# Test 2: Rate limiting only
# Edit config.yaml: enable rate limiting

# Test 3: All features enabled
```

**Expected overhead:**
- Circuit breaker: ~2-5%
- Rate limiting (global): ~5-10%
- Rate limiting (per-IP): ~15-20%
- All features: ~20-25%

### 6. Soak Test (Long Duration)

Test for memory leaks and stability:

```bash
# Run for 4+ hours at moderate load
hey -n 100000000 -c 200 -q 1000 http://localhost:8080

# Monitor memory growth
watch -n 60 'curl -s http://localhost:8080/metrics | grep memory'
```

### 7. Spike Test

Test how system handles sudden load changes:

```bash
# Low load
hey -n 10000 -c 10 http://localhost:8080 &

# Sudden spike
sleep 5
hey -n 100000 -c 1000 http://localhost:8080 &

# Back to normal
sleep 10
hey -n 10000 -c 10 http://localhost:8080
```

## Comparing with Other Proxies

### Test Against nginx

```nginx
# nginx.conf
upstream backend {
    server localhost:8081;
    server localhost:8082;
    server localhost:8083;
}

server {
    listen 9090;
    location / {
        proxy_pass http://backend;
    }
}
```

```bash
# Start nginx
nginx -c /path/to/nginx.conf

# Test ConnX
wrk -t4 -c100 -d30s http://localhost:8080

# Test nginx
wrk -t4 -c100 -d30s http://localhost:9090

# Compare RPS
```

### Test Against HAProxy

```haproxy
# haproxy.cfg
frontend http_front
    bind *:9091
    default_backend http_back

backend http_back
    balance roundrobin
    server backend1 localhost:8081
    server backend2 localhost:8082
    server backend3 localhost:8083
```

```bash
haproxy -f haproxy.cfg
wrk -t4 -c100 -d30s http://localhost:9091
```

### Direct Backend Comparison

```bash
# Test direct to backend (no proxy)
wrk -t4 -c100 -d30s http://localhost:8081

# Calculate proxy overhead
# Overhead% = ((DirectRPS - ProxyRPS) / DirectRPS) * 100
```

## Advanced Profiling

### CPU Profiling

```bash
# Build with profiling
go build -o bin/proxy cmd/proxy/main.go

# Run with CPU profiling
./bin/proxy -config=test/performance/config.yaml -cpuprofile=cpu.prof

# In another terminal, run load test
hey -n 100000 -c 200 http://localhost:8080

# Stop proxy (Ctrl+C) and analyze
go tool pprof cpu.prof
```

### Memory Profiling

```bash
# Add memory profiling flag
./bin/proxy -config=test/performance/config.yaml -memprofile=mem.prof

# Run soak test
hey -n 1000000 -c 200 http://localhost:8080

# Analyze
go tool pprof mem.prof
```

### Real-time Metrics

```bash
# Monitor metrics in real-time
watch -n 1 'curl -s http://localhost:8080/metrics | grep -E "(requests_per_second|active_connections|response_time)"'
```

## Interpreting Results

### Good Indicators
✅ Consistent latency under load
✅ Error rate < 0.01%
✅ Linear scaling with cores
✅ Stable memory usage
✅ <10% proxy overhead

### Warning Signs
⚠️ Increasing latency over time
⚠️ Memory growth (possible leak)
⚠️ High error rate (>1%)
⚠️ Connection timeouts
⚠️ >25% proxy overhead

### Critical Issues
❌ Server crashes under load
❌ Memory leaks
❌ >5% error rate
❌ Latency spikes (>100ms p99)
❌ >50% proxy overhead

## Optimization Tips

1. **Disable rate limiting** for max throughput tests
2. **Use production build** (not debug)
3. **Increase ulimit** for high connection tests
4. **Use separate machines** for client/proxy/backend
5. **Monitor system metrics** (CPU, memory, network)
6. **Warm up** before measuring (discard first 10-30s)
7. **Run multiple times** and take median
8. **Test different payload sizes**
9. **Test HTTP keep-alive** vs new connections
10. **Profile hot paths** with pprof

## Common Issues

### "Too many open files"
```bash
ulimit -n 65536
```

### "Cannot assign requested address"
```bash
sudo sysctl -w net.ipv4.ip_local_port_range="1024 65535"
sudo sysctl -w net.ipv4.tcp_tw_reuse=1
```

### High CPU usage
- Check for inefficient algorithms
- Profile with pprof
- Consider reducing features for max performance

### Memory growth
- Check for goroutine leaks
- Profile with memory profiler
- Review cleanup in Shutdown()

## Reporting Results

When sharing benchmark results, include:

1. **Hardware**: CPU, RAM, OS
2. **Configuration**: Load balancer algorithm, features enabled
3. **Test parameters**: Connections, duration, rate
4. **Results**: RPS, latency (p50/p95/p99), errors
5. **System load**: CPU%, memory usage
6. **Comparison**: vs direct backend, vs nginx/HAProxy

Example:

```
Hardware: Intel i7-9700K (8 cores), 16GB RAM, Ubuntu 22.04
ConnX Config: Round-robin, circuit breaker enabled, rate limiting disabled
Test: wrk -t4 -c200 -d60s
Results:
  - RPS: 87,432
  - p50: 1.8ms
  - p99: 8.3ms
  - Errors: 0.002%
  - CPU: 42%
  - Memory: 85MB
Comparison:
  - Direct backend: 102,341 RPS (14.6% overhead)
  - nginx: 94,127 RPS (ConnX is 7% slower)
  - HAProxy: 91,853 RPS (ConnX is 4% faster)
```
