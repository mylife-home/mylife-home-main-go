DOCKER_REPOSITORY ?= vincenttr
DOCKER_PACKAGE_NAME ?= mylife-home-ui
DOCKER_PACKAGE_VERSION ?= $(shell go run mylife-home-ui/build/print_version.go)
DOCKER_IMAGE_TAG ?= $(DOCKER_REPOSITORY)/$(DOCKER_PACKAGE_NAME):$(DOCKER_PACKAGE_VERSION)
DOCKER_IMAGE_LATEST_TAG ?= $(DOCKER_REPOSITORY)/$(DOCKER_PACKAGE_NAME):latest

OUTPUT_BINARY ?= bin/mylife-home-ui

.PHONY: docker-publish docker-build run prepare generate build

prepare:
	go mod download

build: prepare
	mkdir -p $(dir $(OUTPUT_BINARY))
	CGO_ENABLED=0 go build -o $(OUTPUT_BINARY) mylife-home-ui/main.go
	# go vet -v
	# go test -v

docker-publish: docker-build
	docker push "$(DOCKER_IMAGE_TAG)"
	docker push "$(DOCKER_IMAGE_LATEST_TAG)"

docker-build:
	docker build --pull -t "$(DOCKER_IMAGE_TAG)" -t "$(DOCKER_IMAGE_LATEST_TAG)" -f Dockerfile ..

run:
	go run mylife-home-ui/main.go --log-console
