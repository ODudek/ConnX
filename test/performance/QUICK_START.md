# Performance Testing - Quick Start

## 🚀 1-Minute Quick Test

```bash
# Terminal 1: Start backends
make backends

# Terminal 2: Start ConnX
make run-perf

# Terminal 3: Run basic benchmark
make bench-basic
```

## 📊 Expected Results

On a modern machine (4+ cores, 16GB RAM):

- **Requests/sec**: 50,000 - 100,000+
- **p50 latency**: < 2ms
- **p99 latency**: < 10ms
- **Memory usage**: < 100MB
- **Error rate**: < 0.01%

## 🔥 Full Test Suite

```bash
# Run all benchmarks (takes ~5-10 minutes)
make bench-all
```

This runs:
1. Basic performance test
2. Load test (increasing connections)
3. Algorithm comparison

## 📈 What to Measure

### Key Metrics

1. **Throughput** - Requests per second (RPS)
2. **Latency** - Response time distribution (p50, p95, p99)
3. **Error Rate** - Failed requests percentage
4. **Resource Usage** - CPU, Memory, Connections

### Tools Used

- **wrk** - HTTP benchmarking tool (recommended)
- **hey** - Modern load generator (Go-based)
- **vegeta** - Constant rate testing
- **Apache Bench (ab)** - Simple tests

## 🎯 Common Scenarios

### Test 1: Find Maximum RPS

```bash
make bench-load
```

This gradually increases load to find your limit.

### Test 2: Compare Algorithms

```bash
make bench-compare
```

Tests performance of:
- round-robin
- least-connections
- weighted-round-robin

### Test 3: Stress Test

```bash
make bench-stress
```

Pushes system beyond normal limits (10,000 concurrent connections).

## 📝 Manual Testing

### Using wrk

```bash
# 100 connections for 30 seconds
wrk -t4 -c100 -d30s --latency http://localhost:8080
```

### Using hey

```bash
# 100,000 requests with 200 connections
hey -n 100000 -c 200 http://localhost:8080
```

### Using vegeta

```bash
# 5,000 req/sec for 30 seconds
echo "GET http://localhost:8080" | vegeta attack -rate=5000 -duration=30s | vegeta report
```

## 🔧 Troubleshooting

### "Connection refused"

Make sure backends are running:
```bash
make backends
```

### "Too many open files"

Increase file descriptor limit:
```bash
ulimit -n 65536
```

### Low performance

1. Check CPU usage: `top`
2. Disable rate limiting in config
3. Use production build (not debug)
4. Ensure backends are fast enough

## 📚 Next Steps

- Read [BENCHMARKING.md](BENCHMARKING.md) for detailed guide
- Read [README.md](README.md) for test scenarios
- Compare with nginx/HAProxy
- Profile with pprof for optimization

## 🎓 Tips

1. **Warm up first** - Discard first 10-30 seconds
2. **Run 3-5 times** - Take median result
3. **Monitor system** - Watch CPU, memory, network
4. **Test different configs** - Compare features on/off
5. **Use separate machines** - For realistic results

## Example Output

```
Running 30s test @ http://localhost:8080
  4 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.85ms    2.34ms  45.67ms   89.23%
    Req/Sec    18.6k     2.3k    24.5k    73.00%
  Latency Distribution
     50%    1.23ms
     75%    2.14ms
     90%    4.32ms
     99%    9.87ms
  2,231,456 requests in 30.00s, 345.67MB read
Requests/sec:  74,381.87
Transfer/sec:   11.52MB
```

This shows:
- ✅ 74,381 requests/sec
- ✅ 1.23ms median latency
- ✅ 9.87ms p99 latency
- ✅ 0 errors

**That's excellent performance!** 🎉
