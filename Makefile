.PHONY: generate build build-desktop test dev dev-desktop clean

# Build tags for Wails desktop:
#   dev        = Wails dev mode (use 'production' for release builds)
#   webkit2_41 = Use webkit2gtk-4.1 (Ubuntu 24.04+)
DESKTOP_TAGS ?= dev,webkit2_41

generate:
	templ generate

build: generate
	go build -o bin/sshmasher ./cmd/server

build-desktop: generate
	CGO_ENABLED=1 go build -tags "$(DESKTOP_TAGS)" -o bin/sshmasher-desktop ./cmd/desktop

test:
	go test ./internal/...

dev: generate
	go run ./cmd/server

dev-desktop: generate
	CGO_ENABLED=1 go run -tags "$(DESKTOP_TAGS)" ./cmd/desktop

clean:
	rm -rf bin/ dist/
	find . -name '*_templ.go' -delete
