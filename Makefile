# Audio Recorder Makefile

# Project settings
BINARY_NAME = audio-recorder
GO = go
GOFLAGS = -v
BUILD_DIR = .
CONFIG_FILE = config.yaml

# Platform detection
OS = $(shell uname -s)
ARCH = $(shell uname -m)

# Build targets
.PHONY: all clean build run test install docker docker-build docker-run help

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) .

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	$(GO) clean

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Run with custom config
run-config: build
	@echo "Running $(BINARY_NAME) with custom config..."
	./$(BINARY_NAME) -config $(CONFIG_FILE)

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Install dependencies
install:
	@echo "Installing dependencies..."
	$(GO) mod download

# Build for different platforms
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-linux-amd64 .

build-darwin:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-darwin-amd64 .

build-arm:
	@echo "Building for ARM (Raspberry Pi)..."
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-linux-arm64 .

# Docker targets
docker:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):latest .

docker-build: docker

docker-run:
	@echo "Running Docker container..."
	docker run --rm -it --name $(BINARY_NAME) \
		--device /dev/snd:/dev/snd \
		$(BINARY_NAME):latest

# Cross-compile for Raspberry Pi
pi: build-arm
	@echo "Built for Raspberry Pi: $(BINARY_NAME)-linux-arm64"

# Help target
help:
	@echo "Audio Recorder Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  all          - Build the binary (default)"
	@echo "  build       - Build the binary"
	@echo "  clean       - Clean build artifacts"
	@echo "  run         - Build and run the application"
	@echo "  test        - Run tests"
	@echo "  install     - Install dependencies"
	@echo "  build-linux - Cross-compile for Linux"
	@echo "  build-darwin - Cross-compile for macOS"
	@echo "  build-arm  - Cross-compile for ARM"
	@echo "  docker      - Build Docker image"
	@echo "  docker-run - Run Docker container"
	@echo "  pi         - Build for Raspberry Pi"
	@echo "  help       - Show this help"