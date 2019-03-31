FROM golang:1.12

MAINTAINER clooooode<jackey8616@gmail.com>

EXPOSE 8080

WORKDIR /app

COPY . /app

RUN go mod download

ENTRYPOINT ["go", "run", "main.go"]