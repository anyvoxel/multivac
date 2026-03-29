# The old school Makefile, following are required targets. The Makefile is written
# to allow building multiple binaries. You are free to add more targets or change
# existing implementations, as long as the semantics are preserved.
#
#   make              - default to 'build' target
#   make lint         - code analysis
#   make test         - run unit test (or plus integration test)
#   make build        - alias to build-local target
#   make build-local  - build local binary targets
#   make build-linux  - build linux binary targets
#   make container    - build containers
#   $ docker login registry -u username -p xxxxx
#   make push         - push containers
#   make clean        - clean up targets
#

# This repo's root import path (under GOPATH).
ROOT := github.com/anyvoxel/multivac

# Module name.
NAME := multivac

IMAGE_PREFIX ?= $(strip )
IMAGE_SUFFIX ?= $(strip )

# Container registries.
REGISTRY ?=

export SHELL := /bin/bash
export SHELLOPTS := errexit

CMD_DIR := ./cmd/$(NAME)
OUTPUT_DIR := ./bin
BUILD_DIR := ./build
WEB_DIR := ./web
WEB_EMBED_DIR := ./internal/webui/dist

IMAGE_NAME := $(IMAGE_PREFIX)$(NAME)$(IMAGE_SUFFIX)

VERSION      ?= $(shell git describe --tags --always --dirty)
BRANCH       ?= $(shell git branch | grep \* | cut -d ' ' -f2)
GITCOMMIT    ?= $(shell git rev-parse HEAD)
GITTREESTATE ?= $(if $(shell git status --porcelain),dirty,clean)
BUILDDATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
appVersion   ?= $(VERSION)

CPUS ?= $(shell /bin/bash hack/read_cpus_available.sh)

DOCKER_LABELS ?= git-describe="$(shell date -u +v%Y%m%d)-$(shell git describe --tags --always --dirty)"

GOPATH ?= $(shell go env GOPATH)
BIN_DIR := $(GOPATH)/bin
GOLANGCI_LINT_VERSION := v2.11.4
GOLANGCI_LINT := $(BIN_DIR)/golangci-lint-v2

export GOFLAGS ?= -count=1

.PHONY: lint test build build-local build-linux container push clean

build: build-local

.PHONY: web-build
web-build:
	@cd $(WEB_DIR) && npm install
	@cd $(WEB_DIR) && npm run build
	@rm -rf $(WEB_EMBED_DIR)
	@mkdir -p $(WEB_EMBED_DIR)
	@cp -R $(WEB_DIR)/dist/* $(WEB_EMBED_DIR)/

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) --version
	@$(GOLANGCI_LINT) run -v

$(GOLANGCI_LINT):
	@mkdir -p $(BIN_DIR)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(BIN_DIR) $(GOLANGCI_LINT_VERSION)
	@mv -f $(BIN_DIR)/golangci-lint $(GOLANGCI_LINT)

test:
	@go test -v -race -ldflags="-extldflags=\"-Wl,-segprot,__TEXT,rwx,rx\"" -coverpkg=./... -coverprofile=coverage.out -gcflags="all=-N -l" ./...
	@go tool cover -func coverage.out | tail -n 1 | awk '{ print "Total coverage: " $$3 }'

build-local:
	@$(MAKE) web-build
	@mkdir -p $(OUTPUT_DIR)
	@go build -v -o $(OUTPUT_DIR)/$(NAME) \
	  -ldflags "-s -w -X $(ROOT)/pkg/utils/version.module=$(NAME) \
	    -X $(ROOT)/pkg/utils/version.version=$(VERSION) \
	    -X $(ROOT)/pkg/utils/version.branch=$(BRANCH) \
	    -X $(ROOT)/pkg/utils/version.gitCommit=$(GITCOMMIT) \
	    -X $(ROOT)/pkg/utils/version.gitTreeState=$(GITTREESTATE) \
	    -X $(ROOT)/pkg/utils/version.buildDate=$(BUILDDATE)" \
	  $(CMD_DIR);

build-linux:
	/bin/bash -c 'GOOS=linux GOARCH=amd64 GOPATH=/go GOFLAGS="$(GOFLAGS)" \
	  go build -v -o $(OUTPUT_DIR)/$(NAME) \
	    -ldflags "-s -w -X $(ROOT)/pkg/utils/version.module=$(NAME) \
	      -X $(ROOT)/pkg/utils/version.version=$(VERSION) \
	      -X $(ROOT)/pkg/utils/version.branch=$(BRANCH) \
	      -X $(ROOT)/pkg/utils/version.gitCommit=$(GITCOMMIT) \
	      -X $(ROOT)/pkg/utils/version.gitTreeState=$(GITTREESTATE) \
	      -X $(ROOT)/pkg/utils/version.buildDate=$(BUILDDATE)" \
		$(CMD_DIR)'

container:
	@docker build -t $(REGISTRY)$(IMAGE_NAME):$(VERSION) \
	  --label $(DOCKER_LABELS) \
	  -f $(BUILD_DIR)/Dockerfile .;

push: container
	@docker push $(REGISTRY)/$(IMAGE_NAME):$(VERSION);

clean:
	@-rm -vrf ${OUTPUT_DIR} output coverage.out
