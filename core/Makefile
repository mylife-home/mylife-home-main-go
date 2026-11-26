DOCKER_REPOSITORY ?= vincenttr
DOCKER_PACKAGE_NAME ?= mylife-home-core
DOCKER_PACKAGE_VERSION ?= $(shell go run mylife-home-core/build/print_version.go)
DOCKER_IMAGE_TAG ?= $(DOCKER_REPOSITORY)/$(DOCKER_PACKAGE_NAME):$(DOCKER_PACKAGE_VERSION)
DOCKER_IMAGE_LATEST_TAG ?= $(DOCKER_REPOSITORY)/$(DOCKER_PACKAGE_NAME):latest

APK_PACKAGE_NAME ?= mylife-home-core
APK_PACKAGE_VERSION ?= $(shell go run mylife-home-core/build/print_version.go)
APK_PACKAGE_RELEASE ?= 0
APK_ARCH ?= aarch64
APK_GOARCH ?= arm64
APK_ALPINE_VERSION ?= 3.20
APK_KEY_DIR ?= $(abspath $(CURDIR)/../../mylife-home-studio-data-dev/build-secrets/abuild)
APK_KEY_NAME ?= mylife-home-builder.rsa
APK_REPOSITORY ?= http://rpi-alpine-packages.apps.mti-team2.dyndns.org

OUTPUT_BINARY ?= bin/mylife-home-core

.PHONY: docker-publish docker-build apk-publish apk-build run prepare generate build

prepare:
	go mod download

generate:
	@echo "Generating code for all plugins..."
	@for dir in plugins/*; do \
		if [ -d "$$dir" ]; then \
			echo "  - $$dir"; \
			go generate "$$dir/main.go"; \
		fi \
	done

build: prepare generate
	mkdir -p $(dir $(OUTPUT_BINARY))
	CGO_ENABLED=0 go build -o $(OUTPUT_BINARY) mylife-home-core/main.go
	# go vet -v
	# go test -v

docker-publish: docker-build
	docker push "$(DOCKER_IMAGE_TAG)"
	docker push "$(DOCKER_IMAGE_LATEST_TAG)"

docker-build:
	docker build --pull -t "$(DOCKER_IMAGE_TAG)" -t "$(DOCKER_IMAGE_LATEST_TAG)" -f Dockerfile ..

# abuild only packages the cross-compiled binary, so it runs on the host
# architecture with CARCH set to the target. The public key is trusted inside
# the container because abuild indexes what it just built, and it signs with
# the key the repository and the Pis already trust.
apk-build: prepare generate
	CGO_ENABLED=0 GOOS=linux GOARCH=$(APK_GOARCH) go build -trimpath -ldflags='-s -w' -o alpine/$(APK_PACKAGE_NAME) mylife-home-core/main.go
	docker run --rm \
	  -v "$(CURDIR)/alpine:/pkg" \
	  -v "$(APK_KEY_DIR):/keys:ro" \
	  -e PKGVER=$(APK_PACKAGE_VERSION) \
	  -e PKGREL=$(APK_PACKAGE_RELEASE) \
	  alpine:$(APK_ALPINE_VERSION) sh -ec '\
	    apk add --no-cache abuild >/dev/null; \
	    cp /keys/$(APK_KEY_NAME).pub /etc/apk/keys/; \
	    cd /pkg && rm -rf out; \
	    CARCH=$(APK_ARCH) PACKAGER_PRIVKEY=/keys/$(APK_KEY_NAME) abuild -d -F -P /pkg/out; \
	    chown -R $(shell id -u):$(shell id -g) /pkg/out'

apk-publish: apk-build
	curl --fail --silent --show-error --upload-file alpine/out/$(APK_ARCH)/$(APK_PACKAGE_NAME)-*.apk $(APK_REPOSITORY)/publish

run:
	go run mylife-home-core/main.go --log-console
