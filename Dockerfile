FROM golang:1.25-alpine as builder
RUN apk update && apk add git bash build-base && rm -rf /var/cache/apk/* \
  && mkdir -p /github.com/simagix/hatchet && cd /github.com/simagix \
  && git clone --depth 1 https://github.com/simagix/hatchet.git
WORKDIR /github.com/simagix/hatchet
RUN CGO_CFLAGS="-D_LARGEFILE64_SOURCE" go build -o ./dist/hatchet main/hatchet.go

FROM alpine
LABEL Ken Chen <ken.chen@simagix.com>
RUN addgroup -S simagix && adduser -S simagix -G simagix
USER simagix
WORKDIR /home/simagix
COPY --from=builder /github.com/simagix/hatchet/dist/hatchet /bin/hatchet

CMD ["/bin/sh","-c","sleep infinity"]