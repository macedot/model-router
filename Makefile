.PHONY: build clean run test release

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//')

build:
	@go build -ldflags "-X main.FullVersion=$(VERSION)" -o model-router .
	@echo "Built model-router v$(VERSION)"

run: build
	@./model-router

test:
	go test ./...

clean:
	rm -f model-router

release:
	@if [ "$(VERSION)" = "$(shell git describe --always --dirty 2>/dev/null)" ]; then \
		echo "VERSION is required: make release VERSION=0.5.0"; \
		exit 1; \
	fi
	@git tag -a v$(VERSION) -m "Release v$(VERSION)"
	@git push origin v$(VERSION)
	@gh release create v$(VERSION) --generate-notes
