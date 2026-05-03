BINARY_NAME=ootmm-autotracker
VERSION=0.1.0
OOTMM_REPO?=../OoTMM
GO?=go
BUILD_FLAGS?=-trimpath -buildvcs=false
LDFLAGS?=-s -w

.PHONY: all build build-all clean deps generate-slot-mapping generate-location-mapping generate-special-locations

all: build

deps:
	go mod tidy

generate-slot-mapping:
	python3 tools/generate_inventory_slots.py --ootmm-repo $(OOTMM_REPO) --output ootmm/inventory_slots.json

generate-location-mapping:
	python3 tools/generate_locations.py --ootmm-repo $(OOTMM_REPO) --output ootmm/locations.json

generate-special-locations:
	python3 tools/generate_special_locations.py --ootmm-repo $(OOTMM_REPO) --output ootmm/special_locations.json --oot-output ootmm/oot_special_locations.json

build: deps
	$(GO) build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .

build-all: deps
	mkdir -p dist
	GOOS=linux   GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=windows GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-windows-amd64.exe .
	GOOS=darwin  GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 $(GO) build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-arm64 .

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

run: build
	./$(BINARY_NAME)
