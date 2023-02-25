FROM golang:1.19-alpine as builder
RUN apk update && apk add git bash build-base && rm -rf /var/cache/apk/* \
  && mkdir -p /go/src/github.com/simagix/hatchet
ADD . /go/src/github.com/simagix/hatchet
WORKDIR /go/src/github.com/simagix/hatchet
RUN ./build.sh
FROM alpine
LABEL Ken Chen <ken.chen@simagix.com>
RUN addgroup -S simagix && adduser -S simagix -G simagix
USER simagix
WORKDIR /home/simagix
COPY --from=builder /go/src/github.com/simagix/hatchet/dist/hatchet /hatchet
CMD ["/hatchet", "--version"]
