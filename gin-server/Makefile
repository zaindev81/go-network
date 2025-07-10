BINARY_NAME=tidy

# Basic commands
.PHONY: run
run:
	go run cmd/server/main.go

.PHONY: build
build:
	go build -o $(BINARY_NAME) cmd/server/main.go

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	rm -f $(BINARY_NAME)

# Development
.PHONY: dev
dev:
	SERVER_MODE=debug LOG_LEVEL=debug go run cmd/server/main.go

# Docker
.PHONY: docker
docker:
	docker build -t $(BINARY_NAME) .
	docker run -p 8080:8080 $(BINARY_NAME)

# Utilities
.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: deps
deps:
	go mod tidy

.PHONY: help
help:
	@echo "Commands:"
	@echo "  run    - Run the server"
	@echo "  build  - Build binary"
	@echo "  test   - Run tests"
	@echo "  clean  - Clean build files"
	@echo "  dev    - Run in debug mode"
	@echo "  docker - Build and run with Docker"
	@echo "  fmt    - Format code"
	@echo "  deps   - Tidy dependencies"