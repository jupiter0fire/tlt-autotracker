BINARY_NAME=ootmm-autotracker
VERSION=0.1.0

.PHONY: all build build-all clean deps

all: build

deps:
	go mod tidy

build: deps
	go build -o $(BINARY_NAME) .

build-all: deps
	GOOS=linux   GOARCH=amd64 go build -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build -o dist/$(BINARY_NAME)-windows-amd64.exe .
	GOOS=darwin  GOARCH=amd64 go build -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build -o dist/$(BINARY_NAME)-darwin-arm64 .

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

run: build
	./$(BINARY_NAME)
