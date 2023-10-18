FROM golang:1.21.1 as build

WORKDIR /src
COPY . .

RUN go mod download
RUN go generate mylife-home-core-plugins/driver-absoluta/main.go
RUN go generate mylife-home-core-plugins/logic-base/main.go
RUN go generate mylife-home-core-plugins/logic-clim/main.go
RUN go generate mylife-home-core-plugins/logic-colors/main.go
RUN go generate mylife-home-core-plugins/logic-selectors/main.go
RUN go generate mylife-home-core-plugins/logic-timers/main.go
RUN go generate mylife-home-core-plugins/ui-base/main.go
# RUN go vet -v
# RUN go test -v

RUN CGO_ENABLED=0 go build -o /bin/mylife-home-core mylife-home-core/main.go

FROM gcr.io/distroless/static-debian11

WORKDIR /app
COPY --from=build /bin/mylife-home-core /app
ENTRYPOINT ["/app/mylife-home-core", "--log-console"]
