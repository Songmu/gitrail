VERSION = $(shell godzil show-version)
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X github.com/Songmu/gitrail.revision=$(CURRENT_REVISION)"
u := $(if $(update),-u)

.PHONY: deps
deps:
	go get ${u}
	go mod tidy

.PHONY: devel-deps
devel-deps:
	go install github.com/Songmu/godzil/cmd/godzil@latest
	go install github.com/tcnksm/ghr@latest

.PHONY: test
test:
	go test

.PHONY: build
build:
	go build -ldflags=$(BUILD_LDFLAGS) ./cmd/gitrail

.PHONY: install
install:
	go install -ldflags=$(BUILD_LDFLAGS) ./cmd/gitrail

.PHONY: prepare-release
prepare-release: devel-deps
	go get
	go mod tidy
	godzil credits -w
	git add go.mod go.sum CREDITS

DIST_DIR = dist
.PHONY: crossbuild
crossbuild: devel-deps
	go mod download
	rm -rf $(DIST_DIR)
	godzil crossbuild -pv=v$(VERSION) -build-ldflags=$(BUILD_LDFLAGS) \
      -os=linux,darwin -d=$(DIST_DIR) ./cmd/*
	cd $(DIST_DIR) && shasum -a 256 $$(find * -type f -maxdepth 0) > SHA256SUMS
