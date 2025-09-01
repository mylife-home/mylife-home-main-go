FROM golang:1.22.3 AS build

WORKDIR /src
COPY . .

WORKDIR /src/core
RUN go mod download
RUN go generate plugins/driver-absoluta/main.go
RUN go generate plugins/driver-klf200/main.go
RUN go generate plugins/driver-notifications/main.go
RUN go generate plugins/driver-tahoma/main.go
RUN go generate plugins/logic-base/main.go
RUN go generate plugins/logic-clim/main.go
RUN go generate plugins/logic-colors/main.go
RUN go generate plugins/logic-selectors/main.go
RUN go generate plugins/logic-timers/main.go
RUN go generate plugins/ui-base/main.go
# RUN go vet -v
# RUN go test -v

RUN CGO_ENABLED=0 go build -o /bin/mylife-home-core mylife-home-core/main.go

FROM gcr.io/distroless/static-debian11

WORKDIR /app
COPY --from=build /bin/mylife-home-core /app
ENTRYPOINT ["/app/mylife-home-core", "--log-console"]
