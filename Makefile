.PHONY: build run clean vet tidy

# Default target
build:
	go build -o maild ./cmd/maild/

# Build and run
run: build
	./maild

# Run without rebuilding (if binary already exists)
start:
	./maild

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Tidy dependencies
tidy:
	go mod tidy

# Clean build artifacts
clean:
	rm -f maild

# Build for Linux (static, no libc dependency)
build-static:
	CGO_ENABLED=0 go build -o maild ./cmd/maild/

# Show help
help:
	@echo "make build        - Build the maild binary"
	@echo "make run          - Build and run"
	@echo "make start        - Run existing binary"
	@echo "make fmt          - Format Go code"
	@echo "make vet          - Run go vet"
	@echo "make tidy         - Tidy Go modules"
	@echo "make clean        - Remove build artifacts"
	@echo "make build-static - Build static binary (no libc)"
