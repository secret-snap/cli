BINARY=secretsnap
PLATFORMS=linux/amd64 darwin/amd64 darwin/arm64
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build clean test release install

build:
	@echo "🔨 Building $(BINARY)..."
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/$(BINARY) ./main.go

clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/

test:
	@echo "🧪 Running tests..."
	go test ./...

install: build
	@echo "📦 Installing $(BINARY)..."
	cp bin/$(BINARY) /usr/local/bin/

release: clean
	@echo "🚀 Building releases for $(PLATFORMS)..."
	@for platform in $(PLATFORMS); do \
		IFS='/' read -r GOOS GOARCH <<< "$$platform"; \
		echo "Building for $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/$(BINARY)-$$GOOS-$$GOARCH ./main.go; \
	done

# Development helpers
dev: build
	@echo "🔄 Running in development mode..."
	./bin/$(BINARY)

fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...

vet:
	@echo "🔍 Running go vet..."
	go vet ./...

lint:
	@echo "🔍 Running linter..."
	golangci-lint run

# Help
help:
	@echo "Available targets:"
	@echo "  build     - Build the binary"
	@echo "  clean     - Clean build artifacts"
	@echo "  test      - Run tests"
	@echo "  install   - Install to /usr/local/bin"
	@echo "  release   - Build for all platforms"
	@echo "  dev       - Build and run in development"
	@echo "  fmt       - Format code"
	@echo "  vet       - Run go vet"
	@echo "  lint      - Run linter"
	@echo "  help      - Show this help"
