GOCMD=go
GOBUILD=$(GOCMD) build
BINARY_NAME=./bin/database

all: test build run_server

wire_build:
	wire gen ./cmd/database
	echo "wire build"

build:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/database
	echo "binary build"

run_server:
	LOG_LEVEL=debug $(BINARY_NAME)

test:
	$(GOCMD) test -v ./...
