FROM golang:1.25.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /usr/local/bin/silverfish-backend

# ---

FROM alpine:3.20

LABEL maintainer=clooooode<jackey8616@gmail.com>

RUN apk add --no-cache \
    chromium \
    nss \
    freetype \
    harfbuzz \
    dbus \
    libxml2 \
    libxslt \
    libjpeg-turbo \
    libpng \
    fontconfig \
    glib
RUN which chromium

EXPOSE 10000

COPY --from=builder /usr/local/bin/silverfish-backend /usr/local/bin/silverfish-backend

CMD ["/usr/local/bin/silverfish-backend"]
