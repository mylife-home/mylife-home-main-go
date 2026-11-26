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

### publish

```shell
# Note: version is fetched from `mylife-home-core/pkg/version/value.go`
make apk-publish
```

Cross-compiles for aarch64, packages it with abuild signed by the shared
builder key, and uploads it to the `rpi-alpine-packages` repository. Set
`APK_PACKAGE_RELEASE` to republish the same version as a new package release.

### install on the target

```bash
apk update && apk add mylife-home-core
rc-update add mylife-home-core
rc-service mylife-home-core start
lbu commit -d
```

## TODO

### async issues

emit/dispatch async break the order of messages. eg emit true then false for an action may break the order and send false then true
so we should have a MQTT send/receive queue, which keep order, and remove all async stuff in registry/bus_listener/bus_publisher after