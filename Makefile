.PHONY: build clean run test release

VERSION ?= 0.0.1

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
	@if [ -z "$(VERSION)" ]; then \
		echo "VERSION is required: make release VERSION=1.0.0"; \
		exit 1; \
	fi
	@git tag -a v$(VERSION) -m "Release v$(VERSION)"
	@git push origin v$(VERSION)
	@gh release create v$(VERSION) --generate-notes
