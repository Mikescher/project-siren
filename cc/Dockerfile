FROM golang:1-alpine AS builder

RUN apk add --no-cache tzdata ca-certificates openssl make git tar coreutils bash

COPY . /buildsrc

RUN cd /buildsrc && rm -rf ".swaggobin" && make build



FROM alpine:latest

RUN apk add --no-cache tzdata

COPY --from=builder   /buildsrc/_build/server  /app/server

RUN mkdir /data

WORKDIR /app

EXPOSE 8000

CMD ["/app/server"]
