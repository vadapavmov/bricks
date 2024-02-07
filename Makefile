# Name of the binary to build
BINARY_NAME=bricks

# Go source files
SRC=$(shell find . -name "*.go" -type f)

# Build the binary for the current platform
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/bricks

build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/bricks

build-race:
	CGO_ENABLED=0 go build -race -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/bricks

# Clean the project
clean:
	go clean
	rm -f $(BINARY_NAME)

# Run the tests
test:
	go test -v ./...
