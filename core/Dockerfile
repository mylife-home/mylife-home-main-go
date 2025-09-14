FROM golang:1.24.2 AS build

WORKDIR /src
COPY . .

WORKDIR /src/core
RUN make build OUTPUT_BINARY=/bin/mylife-home-core

FROM gcr.io/distroless/static-debian11

WORKDIR /app
COPY --from=build /bin/mylife-home-core /app
ENTRYPOINT ["/app/mylife-home-core", "--log-console"]
