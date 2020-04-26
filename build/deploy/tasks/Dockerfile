FROM golang:1.12-alpine3.11 AS builder

RUN apk --update upgrade && \
    apk add curl tzdata ca-certificates git make && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /usr/local/bin/migrations ./cmd/migrations/migrations.go

RUN go build -o /usr/local/bin/comic ./cmd/comic/comic.go

RUN go build -o /usr/local/bin/cerebro ./cmd/cerebro/cerebro.go

FROM alpine:3.11 AS tasks

RUN apk add --update ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

COPY --from=builder /usr/local/bin/* /usr/local/bin/
