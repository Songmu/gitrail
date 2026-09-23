CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X github.com/Songmu/gitrail.revision=$(CURRENT_REVISION)"
u := $(if $(update),-u)

.PHONY: deps
deps:
	go get ${u}
	go mod tidy

.PHONY: devel-deps
devel-deps:
	go install github.com/Songmu/gocredits/cmd/gocredits@v0.5.0

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
	go mod tidy
	gocredits -w
	git update-index --add --remove -- go.mod go.sum CREDITS

.PHONY: infra-validate
infra-validate:
	gh infra validate .github/infra.yaml

.PHONY: infra-plan
infra-plan: infra-validate
	gh infra plan .github/infra.yaml

.PHONY: infra-apply
infra-apply: infra-validate
	gh infra apply .github/infra.yaml
