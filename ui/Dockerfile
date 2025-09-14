FROM golang:1.24.2 AS build

# install node
ENV NODE_VERSION=22.19.0
RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash && \
    \. "$HOME/.nvm/nvm.sh" && \
    nvm install ${NODE_VERSION}
ENV PATH="/root/.nvm/versions/node/v${NODE_VERSION}/bin/:${PATH}"

WORKDIR /src
COPY . .

WORKDIR /src/ui
RUN make build OUTPUT_BINARY=/bin/mylife-home-ui

FROM gcr.io/distroless/static-debian11

WORKDIR /app
COPY --from=build /bin/mylife-home-ui /app
COPY --from=build /src/ui/config.docker.yaml /app/config.yaml
ENTRYPOINT ["/app/mylife-home-ui", "--log-console"]
