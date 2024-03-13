# Shortcut targets
default: docker-image

###############################################################################
# Both native and cross architecture builds are supported.
# The target architecture is select by setting the ARCH variable.
# When ARCH is undefined it is set to the detected host architecture.
# When ARCH differs from the host architecture a crossbuild will be performed.
ARCHES = amd64 arm64

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

# unless otherwise set, I am building for my own architecture, i.e. not cross-compiling
ARCH ?= $(BUILDARCH)

###############################################################################

PLATFORMS=linux darwin
VERSION ?= notset
DOCKER_IMAGE_NAME ?= kysor
BINARY_NAME ?= kysor
CGO_ENABLED=0

###############################################################################

GO_VERSION := $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1,2)

commit=$(shell git rev-parse HEAD)
tag=$(shell git describe --tags --exact-match 2>/dev/null)

ldflags := $(LDFLAGS)
ldflags += -X github.com/KYVENetwork/kyve-rdk/tools/kysor/cmd/types.Version=$(VERSION) \
		  -X github.com/KYVENetwork/kyve-rdk/tools/kysor/cmd/types.Commit=$(commit) \
		  -X github.com/KYVENetwork/kyve-rdk/tools/kysor/cmd/types.Tag=$(tag)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -ldflags '$(ldflags)'

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
	@go run $(gofumpt_cmd) -l -w -e .
	@echo "✅ Completed formatting!"

lint:
	@echo "📝 Running linter..."
	@go run $(golangci_lint_cmd) run --timeout=10m  --skip-dirs templates --allow-parallel-runners
	@echo "✅ Completed linting!"

###############################################################################
###  						    Binary building  						    ###
###############################################################################

build: ensure_version
	@echo "🏗️  Building binary..."
	go build -o build/$(BINARY_NAME) $(BUILD_FLAGS) .
	@echo "✅ Completed binary build!"

install: ensure_version
	@echo "🏗️  Installing binary..."
	go install $(BUILD_FLAGS) .
	@echo "✅ Completed binary install!"

release: ensure_version
	@echo "🏗️  Building release binaries..."
	@rm -rf release
	@mkdir release
	@touch release/checksums.txt

	@for platform in $(PLATFORMS); do \
		for arch in $(ARCHES); do \
			echo "🖥️ Building for $$platform-$$arch..."; \
			GOOS=$$platform GOARCH=$$arch go build -o release/$(BINARY_NAME)-$$platform-$$arch $(BUILD_FLAGS) .; \
			cd release; \
			sha256sum $(BINARY_NAME)-$$platform-$$arch >> checksums.txt; \
   			zip -q $(BINARY_NAME)-$$platform-$$arch.zip $(BINARY_NAME)-$$platform-$$arch; \
			cd ..; \
		done; \
	done
	@echo "✅ Completed release build!"

###############################################################################
### 							 	Docker 							 		###
###############################################################################

docker-image:
	@echo "🐳 Building docker image..."
	DOCKER_BUILDKIT=1 docker build --platform linux/$(ARCH) -t $(DOCKER_IMAGE_NAME):$(VERSION) .
	@echo "✅ Completed docker image build!"

docker-run:
	docker run -it --rm $(DOCKER_IMAGE_NAME):$(VERSION)
