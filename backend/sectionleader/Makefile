.PHONY: cli-sectionleader server-sectionleader clean all

# Default target
all: server-sectionleader

server-sectionleader:
	go build -o server-sectionleader ./cmd/httpserver

# Clean build artifacts
clean:
	rm -f server-sectionleader