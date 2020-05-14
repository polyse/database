GOCMD=go
GOBUILD=$(GOCMD) build
BINARY_NAME=bin/database

all: wire_build test build run_server

wire_build:
	cd cmd/database && wire
	echo "wire build"

build:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/database
	echo "binary build"

run_server:
	$(BINARY_NAME)

test:
	$(GOCMD) test -v ./...
