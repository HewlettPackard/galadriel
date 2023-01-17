# golang:1.19.2-alpine3.16
ARG builderimage=golang@sha256:46752c2ee3bd8388608e41362964c84f7a6dffe99d86faeddc82d917740c5968
# alpine:3.16.2
ARG baseimage=alpine@sha256:1304f174557314a7ed9eddb4eab12fed12cb0cd9809e4c28f29af86979a3c870

# Build stage
FROM ${builderimage} as builder
RUN apk add build-base ncurses curl
ADD go.mod /galadriel/go.mod
WORKDIR /galadriel
RUN go mod download
ADD . /galadriel/
RUN make build

# Base image
FROM ${baseimage} AS base
RUN apk --no-cache add dumb-init
RUN mkdir -p /opt/galadriel/bin

# Galadriel Server
FROM base AS galadriel-server
COPY --from=builder /galadriel/bin/galadriel-server /opt/galadriel/bin/galadriel-server
WORKDIR /opt/galadriel
ENTRYPOINT ["/usr/bin/dumb-init", "/opt/galadriel/bin/galadriel-server"]
CMD ["run"]

# Galadriel Harvester
FROM base AS galadriel-harvester
COPY --from=builder /galadriel/bin/galadriel-harvester /opt/galadriel/bin/galadriel-harvester
WORKDIR /opt/galadriel
ENTRYPOINT ["/usr/bin/dumb-init", "/opt/galadriel/bin/galadriel-harvester"]
CMD ["run"]
