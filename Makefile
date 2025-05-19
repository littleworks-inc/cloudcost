# Variables
BINARY_NAME=cloudcost
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS=-ldflags "-X github.com/littleworks-inc/cloudcost/cmd.Version=${VERSION} -X github.com/littleworks-inc/cloudcost/cmd.Commit=${COMMIT} -X github.com/littleworks-inc/cloudcost/cmd.BuildDate=${BUILD_DATE}"

# Main build targets
.PHONY: all build clean test fmt lint run install

all: fmt lint build test

build:
	go build ${LDFLAGS} -o ${BINARY_NAME} .

clean:
	go clean
	rm -f ${BINARY_NAME}

test:
	go test -v ./...

fmt:
	go fmt ./...

lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

run: build
	./${BINARY_NAME}

install: build
	go install ${LDFLAGS}

# Cross-compilation
.PHONY: build-all build-linux build-macos build-windows

build-all: build-linux build-macos build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-linux-amd64 .

build-macos:
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-darwin-amd64 .

build-windows:
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-windows-amd64.exe .