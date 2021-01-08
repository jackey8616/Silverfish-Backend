FROM golang:1.14-alpine

MAINTAINER clooooode<jackey8616@gmail.com>

EXPOSE 8080

WORKDIR /

COPY . /

RUN apk add \
    build-base \
    chromium \
    nss-dev

RUN go version
RUN go mod download

ENTRYPOINT ["env", "config=config.json", "go", "run", "main.go", "config.go"]
