BINARY=claude-orchestrator
CMD=./cmd/claude-orchestrator
INSTALL_DIR=$(HOME)/.local/bin
VERSION ?= local-$(shell git rev-parse --short HEAD 2>/dev/null || echo dev)
LDFLAGS=-X main.version=$(VERSION)

.PHONY: build run test test-race lint install clean

build:
	go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY) $(CMD)

run: build
	./bin/$(BINARY)

test:
	go test ./... -v

test-race:
	go test -race ./...

lint:
	go vet ./...

# Remove antes de copiar: macOS Sequoia mantém cache de assinatura por path
# e mata (SIGKILL) binários novos cujo signature difere do cacheado.
install: build
	mkdir -p $(INSTALL_DIR)
	rm -f $(INSTALL_DIR)/$(BINARY)
	cp bin/$(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "Installed: $(INSTALL_DIR)/$(BINARY) ($(VERSION))"

clean:
	rm -rf bin/
