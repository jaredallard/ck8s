# syntax=docker/dockerfile:1.0.0-experimental
FROM golang:1.13-alpine3.10 AS builder

ENV GO111MODULE on
ENV GOPROXY https://proxy.golang.org

WORKDIR /src/app

# Build deps
RUN apk add --no-cache --update git make bash ca-certificates

# copy over deps
COPY go.mod go.sum ./
RUN go mod download

# copy over the rest
COPY . /src/app

# build binaries
RUN make CGO_ENABLED=0

FROM alpine:3.10
ENTRYPOINT ["/usr/bin/ck8sd"]
RUN apk add --no-cache ca-certificates
COPY --from=builder /src/app/bin/ck8sd /usr/bin/ck8sd