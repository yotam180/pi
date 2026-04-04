BINARY := pi
BUILD_DIR := bin
INSTALL_DIR := /usr/local/bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/vyper-tooling/pi/internal/cli.version=$(VERSION)"

.PHONY: build test install clean snapshot

# Local dev build. Production releases use GoReleaser (see .goreleaser.yaml).
build:
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/pi

test:
	go test ./... -race

install: build
	cp $(BUILD_DIR)/$(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "Installed $(BINARY) to $(INSTALL_DIR)/$(BINARY)"

clean:
	rm -rf $(BUILD_DIR) dist/

snapshot:
	goreleaser release --snapshot --clean
