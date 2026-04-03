.PHONY: build clean bump run test install release

VERSION ?= 0.0.1
BUILD_FILE := build.txt

get_build = $(shell cat $(BUILD_FILE) 2>/dev/null || echo "0")
next_build = $(shell echo $$(($(get_build) + 1)))
FULL_VERSION := $(VERSION)-b$(next_build)

bump:
	@echo $(next_build) > $(BUILD_FILE)
	@echo "Build: $(next_build) | Full version: $(FULL_VERSION)"

build:
	@go build -ldflags "-X main.FullVersion=$(FULL_VERSION)" -o model-router . && \
		$(MAKE) bump
	@echo "Built model-router v$(FULL_VERSION)"

run: build
	@./model-router

model-router:
	@go build -ldflags "-X main.FullVersion=$(FULL_VERSION)" -o model-router .

test:
	go test ./...

install:
	@bash ./install.sh

clean:
	rm -f model-router build.txt

release:
	@if [ -z "$(VERSION)" ]; then \
		echo "VERSION is required: make release VERSION=1.0.0"; \
		exit 1; \
	fi
	@git tag -a v$(VERSION) -m "Release v$(VERSION)"
	@git push origin v$(VERSION)
	@gh release create v$(VERSION) --generate-notes
