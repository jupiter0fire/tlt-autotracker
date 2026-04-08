BINARY_NAME=ootmm-autotracker
VERSION=0.1.0
OOTMM_REPO?=../OoTMM

.PHONY: all build build-all clean deps generate-slot-mapping

all: build

deps:
	go mod tidy

generate-slot-mapping:
	python3 tools/generate_inventory_slots.py --ootmm-repo $(OOTMM_REPO) --output ootmm/inventory_slots.json

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
