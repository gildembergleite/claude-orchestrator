BINARY=claude-orchestrator
CMD=./cmd/claude-orchestrator

.PHONY: build run test lint install clean

build:
	go build -o bin/$(BINARY) $(CMD)

run: build
	./bin/$(BINARY)

test:
	go test ./... -v

lint:
	go vet ./...

install: build
	cp bin/$(BINARY) ~/.local/bin/$(BINARY)

clean:
	rm -rf bin/
