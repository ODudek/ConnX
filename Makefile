.PHONY: build test run clean bench bench-basic bench-load bench-stress bench-compare bench-setup

build:
	go build -o bin/proxy cmd/proxy/main.go

test:
	go test ./...

run:
	go run cmd/proxy/main.go

clean:
	rm -rf bin/

# Performance testing
bench-setup:
	@echo "Setting up performance test environment..."
	@mkdir -p test/performance/results
	@chmod +x test/performance/*.sh test/performance/*.py
	@echo "✅ Setup complete"

bench-basic: bench-setup
	@echo "Running basic performance test..."
	@./test/performance/run_basic_test.sh

bench-load: bench-setup
	@echo "Running load test..."
	@./test/performance/load_test.sh

bench-stress: bench-setup
	@echo "Running stress test..."
	@./test/performance/stress_test.sh

bench-compare: bench-setup
	@echo "Comparing load balancing algorithms..."
	@./test/performance/compare_algorithms.sh

bench-all: bench-basic bench-load bench-compare
	@echo "All benchmarks completed!"

# Start test backends
backends:
	@echo "Starting test backends on ports 8081, 8082, 8083..."
	@python3 test/performance/simple_backend.py 8081 &
	@python3 test/performance/simple_backend.py 8082 &
	@python3 test/performance/simple_backend.py 8083 &
	@sleep 1
	@echo "✅ Backends started"

# Run proxy with performance config
run-perf: build
	./bin/proxy -config=test/performance/config.yaml

