.PHONY: build

-include .env

# Package
NAME            = webhook
VERSION         = 0.2.0
DESCRIPTION     = Webhook Server (Go/Wasm)
ARCH            ?= amd64 #arm64

# Git
GIT_BRANCH      = $(shell git rev-parse --abbrev-ref HEAD)
GIT_HASH        = $(shell git rev-parse --short HEAD)

# Image
REGISTRY        ?= docker.io
REGISTRY_REPO   ?= ezconnect
DOCKERFILE      ?= Dockerfile
RELEASE_NAME    = $(NAME)

# Build flags
GO_FLAGS					?= CGO_ENABLED=1
GO_LD_FLAGS				?= -X main.Branch=$(GIT_BRANCH) \
	-X main.Version=$(VERSION) \
	-X main.Hash=$(GIT_HASH) \
	-X main.BuildDate=$(shell date +%Y-%m-%d)

GO_DEBUG_FLAGS		?= -ldflags="$(GO_LD_FLAGS) \
	-X $(PACKAGE)/cmd.BuildMode=debug"
GO_RELEASE_FLAGS	?= -ldflags="-s -w $(GO_LD_FLAGS) \
	-X $(PACKAGE)/cmd.BuildMode=production" \
	-trimpath

#: list all targets
help:
	@grep -B1 -E "^[a-zA-Z0-9_%-]+:([^\=]|$$)" Makefile \
		| grep -v -- -- \
		| sed 'N;s/\n/###/' \
		| sed -n 's/^#: \(.*\)###\(.*\):.*/\2###\1/p' \
		| column -t -s '###'

###############################################################################
# Scripts
###############################################################################
#: initialize project
init:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	mise use -g watchexec

#: gofmt (reformat) package sources
fmt:
	go fmt ./...

#: smart, fast linters
lint:
	golangci-lint run ./...

fix:
	go fix ./...

dev:
	watchexec -r -e go,ts -- make run

data:
	mkdir -p build/config build/plugins
	cp default.config.toml build/config/

run: data
	go build -o build/webhook $(GO_DEBUG_FLAGS)
	cd build; ./webhook -c config/default.config.toml

#: Build all binaries and plugins
build:
	@echo "Building webhook server..."
	$(GO_FLAGS) go build -o build/webhook $(GO_RELEASE_FLAGS)

	# @echo "Building Go Echo plugin..."
	# cd plugins/echo && make build

	# @echo "Building Bun (AssemblyScript) Echo plugin..."
	# cd plugins/echo-bun && bun run build
	# @echo "Exporting init, get, run, and destroy methods for Bun plugin..."

#: Clean outputs
clean:
	rm -rf build/

###############################################################################
# OCI
###############################################################################
#: build the image
oci:
	@$(foreach arch,$(ARCH), \
		echo "build: $(RELEASE_NAME):$(VERSION)-$(arch)"; \
		podman build -t $(RELEASE_NAME):$(VERSION)-$(arch) -f $(DOCKERFILE) \
			--arch $(arch) \
			--build-arg arch=$(arch) \
			--annotation org.opencontainers.image.created="$(shell date -I'seconds')" \
			--annotation org.opencontainers.image.description="$(DESCRIPTION)"; \
	)

#: push an image to a specified location that defined in '.env'
oci-push:
	podman login $(REGISTRY) --authfile ~/.config/containers/auth.json

	-podman manifest rm $(RELEASE_NAME):$(VERSION)
	podman manifest create $(RELEASE_NAME):$(VERSION)

	@$(foreach arch,$(ARCH), \
		echo "push: $(REGISTRY)/$(REGISTRY_REPO)/$(RELEASE_NAME):$(VERSION)-$(arch)"; \
		podman push $(RELEASE_NAME):$(VERSION)-$(arch) \
			$(REGISTRY)/$(REGISTRY_REPO)/$(RELEASE_NAME):$(VERSION)-$(arch); \
		podman manifest add $(RELEASE_NAME):$(VERSION) \
			$(REGISTRY)/$(REGISTRY_REPO)/$(RELEASE_NAME):$(VERSION)-$(arch); \
	)

	@echo "push: $(REGISTRY)/$(REGISTRY_REPO)/$(RELEASE_NAME):$(VERSION)"
	podman manifest push $(RELEASE_NAME):$(VERSION) $(REGISTRY)/$(REGISTRY_REPO)/$(RELEASE_NAME):$(VERSION)
