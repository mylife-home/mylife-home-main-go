# mylife-home-core-go

MyLife Home Core, Golang implementation

## generate

```shell
go generate mylife-home-core-plugins-logic-base/main.go 
go generate mylife-home-core-plugins-driver-absoluta/main.go 

```

## run

```shell
go run mylife-home-core/main.go --log-console
```

## TODO

- test bindings
  - normal ops
  - source disconnects
  - target disconnects
  - binding types does not match
  - source connects after target
  - target connects after source
- implements core plugins