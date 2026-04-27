BINARY=claude-orchestrator
CMD=./cmd/claude-orchestrator
INSTALL_DIR=$(HOME)/.local/bin
VERSION ?= local-$(shell git rev-parse --short HEAD 2>/dev/null || echo dev)
LDFLAGS=-X main.version=$(VERSION)

.PHONY: build run test test-race lint install clean release

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

# Cuts a release: validates, creates annotated tag, pushes, and watches
# the goreleaser workflow until success/failure.
#   make release VERSION=v0.3.0
#   make release VERSION=v0.3.0 TAG_MSG="Release v0.3.0 — feature X"
TAG_MSG ?= Release $(VERSION)
release:
	@test -n "$(VERSION)" || (echo "VERSION required: make release VERSION=vX.Y.Z"; exit 1)
	@echo "$(VERSION)" | grep -qE "^v[0-9]+\.[0-9]+\.[0-9]+$$" || (echo "VERSION must be vX.Y.Z (got: $(VERSION))"; exit 1)
	@if [ -n "$$(git status --porcelain)" ]; then echo "Working tree not clean; commit or stash first."; exit 1; fi
	@if [ "$$(git branch --show-current)" != "main" ]; then echo "Must be on main branch (got: $$(git branch --show-current))"; exit 1; fi
	@if git tag | grep -qx "$(VERSION)"; then echo "Tag $(VERSION) already exists locally"; exit 1; fi
	@echo "Creating annotated tag $(VERSION)..."
	git tag -a $(VERSION) -m "$(TAG_MSG)"
	git push origin $(VERSION)
	@echo "Tag pushed. Aguardando workflow disparar..."
	@sleep 5
	@RUN_ID=$$(gh run list --workflow=release.yml --limit 1 --json databaseId --jq '.[0].databaseId'); \
		echo "Watching run $$RUN_ID"; \
		gh run watch $$RUN_ID --exit-status
	@echo "Release published: https://github.com/gildembergleite/claude-orchestrator/releases/tag/$(VERSION)"
