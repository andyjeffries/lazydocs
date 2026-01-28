.PHONY: build install clean test run

BINARY_NAME=lazydocs
BUILD_DIR=./cmd/lazydocs
BUILD_FLAGS=-tags sqlite_fts5
INSTALL_DIR=$(HOME)/.local/bin

# Version info
VERSION?=0.1.0
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

build:
	go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME) $(BUILD_DIR)

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY_NAME) $(INSTALL_DIR)/

uninstall:
	rm -f $(INSTALL_DIR)/$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)
	go clean

test:
	go test $(BUILD_FLAGS) ./...

run: build
	./$(BINARY_NAME)

# Development helpers
deps:
	go mod download
	go mod tidy

fmt:
	go fmt ./...

lint:
	go vet ./...
