# Copyright 2022-present Kuei-chun Chen. All rights reserved.
FROM golang:1.25-alpine AS builder

# Install dependencies first (cached layer)
RUN apk update && apk add git bash && rm -rf /var/cache/apk/*

WORKDIR /github.com/simagix/hatchet

# Copy go.mod and go.sum first for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build for native platform (CGO requires native compilation)
RUN go mod tidy && ./build.sh binary

FROM alpine:3.19
LABEL maintainer="Ken Chen <ken.chen@simagix.com>"
RUN apk add --no-cache ca-certificates
RUN addgroup -S simagix && adduser -S simagix -G simagix
COPY --from=builder /github.com/simagix/hatchet/hatchet /bin/hatchet
RUN ln -s /bin/hatchet /hatchet
USER simagix
WORKDIR /home/simagix

CMD ["/bin/hatchet", "-version"]
