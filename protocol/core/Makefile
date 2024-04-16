# Shortcut targets
default: docker-image

###############################################################################
DOCKERFILE ?= Dockerfile
VERSION ?= notset
DOCKER_IMAGE_NAME=protocol-core

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