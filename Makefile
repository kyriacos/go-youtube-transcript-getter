CLI_PATH := ./cmd/youtube-transcript-getter
BINARY := youtube-transcript-getter
BIN_DIR := bin

.PHONY: build test install release clean

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) $(CLI_PATH)

test:
	go test ./...

install:
	go install $(CLI_PATH)

release:
	@if [ -z "$(VERSION)" ]; then echo "Usage: make release VERSION=v1.0.0"; exit 1; fi
	@if [ -n "$$(git status --porcelain)" ]; then echo "Working tree is not clean. Commit or stash changes first."; exit 1; fi
	$(MAKE) test
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

clean:
	rm -rf $(BIN_DIR)
