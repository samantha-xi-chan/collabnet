# Makefile

.PHONY: all build run clean test

BINARY_NAME=myapp
MAIN_FILE=main.go

all: build

build:
	go build -o $(BINARY_NAME) $(MAIN_FILE)

run:
	go run $(MAIN_FILE)

clean:
	go clean
	rm -f $(BINARY_NAME)

test:
	go test ./...
