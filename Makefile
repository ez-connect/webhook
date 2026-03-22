# Get the value from the package.json
define get_package
$(shell yq -r ".$1" packages/server/package.json)
endef

# Package
NAME					= webhook
VERSION				?= $(call get_package,version)
DESCRIPTION		= $(call get_package,description)
ARCH					?= amd64 #arm64

# Git
GIT_BRANCH		= $(shell git rev-parse --abbrev-ref HEAD)
GIT_HASH			= $(shell git rev-parse --short HEAD)

# Image
REGISTRY			?= docker.io
REGISTRY_REPO	?= ezconnect
DOCKERFILE		?= Dockerfile
RELEASE_NAME	= $(NAME)

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
dev:
	bun run dev

fmt:
	bun run fmt

lint:
	bun run lint

check:
	bun run check

build:
	bun run build

###############################################################################
# OCI
###############################################################################
#: build the image
oci:
	@$(foreach arch,$(ARCH), \
		echo "build: $(RELEASE_NAME):$(VERSION)-$(arch)"; \
		podman build -t $(RELEASE_NAME):$(VERSION)-$(arch) -f $(DOCKERFILE) \
			--arch $(arch) $(args) \
			--build-arg arch=$(arch) \
			--annotation org.opencontainers.image.created="$(shell date -I'seconds')" \
			--annotation org.opencontainers.image.description="$(DESCRIPTION)" \
			--annotation io.artifacthub.package.readme-url="$(README)"; \
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
