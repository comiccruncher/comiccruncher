FROM golang:1.11-alpine3.8 AS builder

RUN apk --update upgrade && \
    apk add curl tzdata ca-certificates git make && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

ENV GOPATH /go/
ENV CGO_ENABLED=0

WORKDIR /app

COPY . .

RUN cp .netrc /root/.netrc

RUN go mod download

RUN rm -rf /root/.netrc

RUN go build -o /usr/local/bin/webapp ./cmd/web/web.go

RUN rm -rf /root/.netrc

FROM alpine:3.8 AS app

RUN apk add --update ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

COPY --from=builder /usr/local/bin/webapp /usr/local/bin/webapp

CMD webapp start -p 8001

EXPOSE 8001
