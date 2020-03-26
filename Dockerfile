FROM golang:1.14

MAINTAINER clooooode<jackey8616@gmail.com>

EXPOSE 8080

WORKDIR /

COPY . /

RUN go version
RUN go mod download

ENTRYPOINT ["env", "config=config.json", "go", "run", "main.go", "config.go"]
