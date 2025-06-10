# Binary name
BINARY_NAME=sectionleader

# Source files
SRC=$(shell find . -name '*.go' -not -path "./vendor/*")

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	go build -o $(BINARY_NAME) ./main

# Run the application
.PHONY: run
run:
	go run ./cmd/sectionleader

# Clean up binaries
.PHONY: clean
clean:
	rm $(BINARY_NAME)
	rm -rf data

# Format the code
.PHONY: fmt
fmt:
	go fmt ./...

# Run tests
.PHONY: test
test:
	go test ./...

# Install dependencies
.PHONY: deps
deps:
	go mod tidy

# Watch for changes and rerun (requires fswatch or similar)
.PHONY: watch
watch:
	fswatch -o $(SRC) | xargs -n1 -I{} make run