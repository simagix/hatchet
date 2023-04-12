FROM golang:1.19-alpine as builder
RUN apk update && apk add git bash build-base && rm -rf /var/cache/apk/* \
  && mkdir -p /github.com/simagix/hatchet && cd /github.com/simagix \
  && git clone --depth 1 https://github.com/simagix/hatchet.git
WORKDIR /github.com/simagix/hatchet
RUN ./build.sh
FROM alpine
LABEL Ken Chen <ken.chen@simagix.com>
RUN addgroup -S simagix && adduser -S simagix -G simagix
USER simagix
WORKDIR /home/simagix
COPY --from=builder /github.com/simagix/hatchet/dist/hatchet /hatchet
CMD ["/hatchet", "--version"]
