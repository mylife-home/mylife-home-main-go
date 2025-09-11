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
apk add mylife-home-core-go

vi /etc/mylife-home/config.yaml
  your-should-override-this => localhost
  supportsBindings => true

rc-service mylife-home-core start
```

## TODO

### async issues

emit/dispatch async break the order of messages. eg emit true then false for an action may break the order and send false then true
so we should have a MQTT send/receive queue, which keep order, and remove all async stuff in registry/bus_listener/bus_publisher after