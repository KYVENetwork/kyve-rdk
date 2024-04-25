# Shortcut targets
default: docker-image

###############################################################################
# Both native and cross architecture builds are supported.
# The target architecture is select by setting the ARCH variable.
# When ARCH is undefined it is set to the detected host architecture.
# When ARCH differs from the host architecture a crossbuild will be performed.
ARCHES = amd64 armv7 arm64

GOOS ?= $(shell uname | tr '[:upper:]' '[:lower:]')

# BUILDARCH is the host architecture
# ARCH is the target architecture
# we need to keep track of them separately
BUILDARCH ?= $(shell uname -m)

# canonicalized names for host architecture
ifeq ($(BUILDARCH),aarch64)
	BUILDARCH=arm64
endif
ifeq ($(BUILDARCH),x86_64)
	BUILDARCH=amd64
endif
ifeq ($(BUILDARCH),armv7l)
	BUILDARCH=armv7
endif

# unless otherwise set, I am building for my own architecture, i.e. not cross-compiling
ARCH ?= $(BUILDARCH)

###############################################################################

GO_VERSION := $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1,2)
DOCKERFILE ?= Dockerfile
VERSION ?= notset
DOCKER_IMAGE_NAME=tendermint-bsync-go
BINARY_NAME ?= tendermint-bsync-go-$(GOOS)-$(ARCH)
CGO_ENABLED=0

###############################################################################

ldflags := $(LDFLAGS)
ldflags += -X github.com/KYVENetwork/kyve-rdk/runtime/tendermint-bsync-go/server.version=$(VERSION)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -ldflags '$(ldflags)'

.PHONY: ensure_version build format lint test test-e2e docker-image docker-run

###############################################################################
###                                 Checks                                  ###
###############################################################################

ensure_version:
ifneq ($(GO_VERSION),1.22)
	$(error ❌  Please run Go v1.22.x..)
endif

###############################################################################
###                          Formatting & Linting                           ###
###############################################################################

gofumpt_cmd=mvdan.cc/gofumpt
golangci_lint_cmd=github.com/golangci/golangci-lint/cmd/golangci-lint

format:
	@echo "📝 Running formatter..."
	@go run $(golangci_lint_cmd) run --timeout=10m
	@echo "✅ Completed formatting!"

lint:
	@echo "📝 Running linter..."
	@go run $(golangci_lint_cmd) run --timeout=10m --allow-parallel-runners
	@echo "✅ Completed linting!"

###############################################################################
### 						 	   Testing									###
###############################################################################

test:
	@echo "🧪 Running tests..."
	@./scripts/test.sh
	@echo "✅ Completed tests!"

test-e2e:
	@echo "🧪 Running end-to-end tests..."
	@../../test/e2e/run-e2e-tests.sh "-test.run '/tendermint-bsync-go\$$'"
	@echo "✅ Completed end-to-end tests!"

###############################################################################
###  						 Building the binary							###
###############################################################################

build: ensure_version
	@echo "⚙️  Building binary..."
	go build -o build/$(BINARY_NAME) $(BUILD_FLAGS) .
	@echo "✅ Completed binary build!"

run:
	./build/$(BINARY_NAME) start --port $(PORT) --debug $(DEBUG)

serve: ensure_version
	go run main.go start --port $(PORT) --debug $(DEBUG)

###############################################################################
### 							 	Docker 							 		###
###############################################################################

docker-image:
	@echo "🏗️  Building docker image..."
	DOCKER_BUILDKIT=1 docker build --platform linux/$(ARCH) -t $(DOCKER_IMAGE_NAME):$(VERSION) --build-arg VERSION=$(VERSION) .
	@echo "✅ Completed docker image build!"

docker-run:
	docker run -it --rm -p 50051:50051 $(DOCKER_IMAGE_NAME):$(VERSION)