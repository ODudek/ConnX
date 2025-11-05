# Performance Testing Guide for ConnX

## Quick Start

```bash
# 1. Start ConnX with test config
./bin/proxy -config=configs/test/performance.yaml

# 2. Start backend servers (in separate terminals)
python3 test/performance/simple_backend.py 8081
python3 test/performance/simple_backend.py 8082
python3 test/performance/simple_backend.py 8083

# 3. Run basic performance test
./test/performance/run_basic_test.sh

# 4. Run full benchmark suite
./test/performance/run_all_benchmarks.sh
```

## Test Tools

### Primary Tools
1. **wrk2** - Constant throughput load testing (recommended)
2. **vegeta** - Constant rate HTTP load testing (Go-based)
3. **hey** - Modern HTTP load generator (Go-based)
4. **Apache Bench (ab)** - Simple, quick tests

### Monitoring Tools
1. **Prometheus** - Scrape /metrics endpoint
2. **Grafana** - Visualize metrics
3. **pprof** - CPU/memory profiling

## Test Scenarios

### 1. Baseline Test
Measure baseline performance with low load.

### 2. Load Test
Gradually increase load to find breaking point.

### 3. Stress Test
Push system beyond normal capacity.

### 4. Spike Test
Sudden load increase/decrease.

### 5. Soak Test
Long duration test (hours) for memory leaks.

### 6. Feature Comparison
Compare performance of different load balancing algorithms.

## Metrics to Track

### From ConnX /metrics endpoint:
- `connx_requests_total` - Total requests
- `connx_response_time_seconds` - Response times
- `connx_errors_total` - Error count
- `connx_active_connections` - Active connections
- `connx_requests_per_second` - Current RPS

### From test tools:
- Requests per second achieved
- Latency distribution (p50, p95, p99)
- Error rate
- Connection errors
- Timeout rate

### System metrics:
- CPU usage (%)
- Memory usage (MB)
- Network I/O (MB/s)
- Open file descriptors
- Context switches

## Expected Performance Targets

Based on gnet framework benchmarks:

- **Simple proxy**: 50,000-100,000+ RPS (depends on hardware)
- **With rate limiting**: ~80% of baseline
- **With circuit breaker**: ~90% of baseline
- **p99 latency**: <10ms (for local backends)
- **Memory**: <200MB under load
- **CPU**: <50% on 4 cores at 50k RPS

## Tips

1. Run tests from separate machine to avoid localhost bias
2. Use multiple backend servers for realistic scenarios
3. Warm up before measuring (first 10-30s)
4. Run each test 3-5 times, take median
5. Monitor both ConnX and backend servers
6. Test different payload sizes
7. Test keep-alive vs new connections
8. Disable debug logging for performance tests
9. Use production-like configuration

## Comparing Results

Compare ConnX against:
- Direct backend access (no proxy)
- nginx (industry standard)
- HAProxy (industry standard)
- Other Go proxies (Traefik, etc.)

Calculate overhead:
```
Overhead % = ((DirectRPS - ProxyRPS) / DirectRPS) * 100
```

Good proxy should have <10% overhead for simple scenarios.
