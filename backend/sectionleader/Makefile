.PHONY: cli-sectionleader server-sectionleader clean all

# Default target
all: cli-sectionleader

# Build CLI version
cli-sectionleader:
	go build -o cli-sectionleader ./cmd/cli

# Build server version (future implementation)
server-sectionleader:
	go build -o server-sectionleader ./cmd/httpserver

# Clean build artifacts
clean:
	rm -f cli-sectionleader server-sectionleader