.PHONY: build test run

build:
	go build -o build/nautikit ./cmd/nautikit/

test:
	go test ./...

run:
	go run ./cmd/nautikit/
