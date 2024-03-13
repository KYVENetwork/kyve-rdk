# Shortcut targets
default: docker-image

###############################################################################
# Both native and cross architecture builds are supported.
# The target architecture is select by setting the ARCH variable.
# When ARCH is undefined it is set to the detected host architecture.
# When ARCH differs from the host architecture a crossbuild will be performed.
ARCHES = amd64 armv7 arm64

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

# set the node architecture for pkg
ifeq ($(ARCH),amd64)
	NODE_ARCH=x64
endif
ifeq ($(ARCH),armv7)
	NODE_ARCH=armv7
endif
ifeq ($(ARCH),arm64)
	NODE_ARCH=arm64
endif

###############################################################################
DOCKERFILE ?= Dockerfile
VERSION ?= notset
DOCKER_IMAGE_NAME=tendermint

###############################################################################
###                          Formatting & Linting                           ###
###############################################################################

format:
	@echo "📝 Running formatter..."
	@yarn format
	@echo "✅ Completed formatting!"

###############################################################################
### 							 	Docker 							 		###
###############################################################################

docker-image:
	@echo "🐳 Building Docker image..."
	@docker build -t $(DOCKER_IMAGE_NAME):$(VERSION) -f $(DOCKERFILE) --build-arg VERSION=$(VERSION) .
	@echo "✅ Completed Docker image build!"