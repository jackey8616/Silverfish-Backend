FROM golang:1.14-alpine

LABEL maintainer=clooooode<jackey8616@gmail.com>

EXPOSE 8080

WORKDIR /

COPY router /router
COPY silverfish /silverfish
COPY config.go main.go go.mod go.sum /

RUN apk add \
    build-base \
    chromium \
    nss-dev

RUN go version
RUN go mod download

ENTRYPOINT ["env", "config=config.json", "go", "run", "main.go", "config.go"]
