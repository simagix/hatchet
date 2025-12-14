FROM golang:1.25-alpine AS builder
RUN apk update && apk add git bash build-base && rm -rf /var/cache/apk/*
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-X main.version=v$(cat version)-$(date +%Y%m%d) -X main.repo=simagix/hatchet" -o ./dist/hatchet main/hatchet.go

FROM alpine
LABEL maintainer="Ken Chen <ken.chen@simagix.com>"
RUN addgroup -S simagix && adduser -S simagix -G simagix
USER simagix
WORKDIR /home/simagix
COPY --from=builder /build/dist/hatchet /bin/hatchet

CMD ["/bin/hatchet", "-version"]