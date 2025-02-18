.PHONY: build test run clean

build:
	go build -o bin/proxy cmd/proxy/main.go

test:
	go test ./...

run:
	go run cmd/proxy/main.go

clean:
	rm -rf bin/
