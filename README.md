# mylife-home-core-go

MyLife Home Core, Golang implementation

## Bump

- common: `mylife-home-common/version/value.go`
- core: `mylife-home-core/pkg/version/value.go`
- plugins: `mylife-home-core-plugins/*/main.go`

## New plugin

- new folder in `mylife-home-core-plugins/`
- Add `Generate` doc below
- Add generate in `Dockerfile`
- Add plugin in `mylife-home-core/main.go`

## Generate

```shell
go generate mylife-home-core-plugins/driver-absoluta/main.go
go generate mylife-home-core-plugins/driver-tahoma/main.go
go generate mylife-home-core-plugins/logic-base/main.go
go generate mylife-home-core-plugins/logic-clim/main.go
go generate mylife-home-core-plugins/logic-colors/main.go
go generate mylife-home-core-plugins/logic-selectors/main.go
go generate mylife-home-core-plugins/logic-timers/main.go
go generate mylife-home-core-plugins/ui-base/main.go
```

## Run

```shell
go run mylife-home-core/main.go --log-console
```

## Publish

```shell
# TODO: update version
bash
export DOCKER_IMAGE_TAG="vincenttr/mylife-home-core:go-1.0.4"
docker build --pull -t "$DOCKER_IMAGE_TAG" .
docker push "$DOCKER_IMAGE_TAG"
exit
```