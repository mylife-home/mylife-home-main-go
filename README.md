# mylife-home-core-go

MyLife Home Core, Golang implementation

## Bump

- common: `mylife-home-common/version/value.go`
- core: `mylife-home-core/pkg/version/value.go`
- plugins: `mylife-home-core-plugins/*/main.go`
- update version number in [publish below](#publish)
- [release github](https://github.com/mylife-home/mylife-home-core-go/releases)

## New plugin

- new folder in `mylife-home-core-plugins/`
- Add `Generate` doc below
- Add generate in `Dockerfile`
- Add plugin in `mylife-home-core/main.go`
- Add plugin in `rpi-alpine-build APKBUILD`

## Generate

```shell
go generate mylife-home-core-plugins/driver-absoluta/main.go
go generate mylife-home-core-plugins/driver-klf200/main.go
go generate mylife-home-core-plugins/driver-notifications/main.go
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

## Docker

### publish

```shell
# TODO: update version
bash
export DOCKER_IMAGE_TAG="vincenttr/mylife-home-core:go-1.0.13"
docker build --pull -t "$DOCKER_IMAGE_TAG" . && docker push "$DOCKER_IMAGE_TAG"
exit
```

### Investigate last crash

```bash
kubectl logs -n mylife-home pod-xxx -p
```

## Alpine - Raspberry PI

- publish using rpi-alpine-build
- test on rpi:

```bash
rc-service mylife-home-core stop

apk del mylife-home-core mylife-home-core-plugins-logic-selectors mylife-home-core-plugins-logic-colors mylife-home-core-plugins-logic-timers mylife-home-core-plugins-logic-base mylife-home-core-plugins-ui-base mylife-home-core-plugins-driver-mpd mylife-home-core-plugins-driver-absoluta mylife-home-core-plugins-logic-clim mylife-home-core-plugins-driver-tahoma mylife-home-core-plugins-driver-broadlink

apk add mylife-home-core-go

vi /etc/mylife-home/config.yaml
  your-should-override-this => localhost
  supportsBindings => true

rc-service mylife-home-core start
```
