# Core

```shell
cd core
```

## New plugin

- new folder in `mylife-home-core-plugins/`
- Add plugin in `mylife-home-core/main.go`

## Generate plugins metadata

```shell
make generate
```

## Run

```shell
make run
```

## Docker

### publish

```shell
# Note: version is fetched from `mylife-home-core/pkg/version/value.go`
make docker-publish
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
