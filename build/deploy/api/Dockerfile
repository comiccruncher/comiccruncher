FROM golang:1.10-alpine3.8 AS builder

RUN apk --update upgrade && \
    apk add curl tzdata ca-certificates git make && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

RUN curl https://raw.githubusercontent.com/golang/dep/v0.5.0/install.sh | sh

ENV GOPATH /go/

WORKDIR /go/src/github.com/comiccruncher/comiccruncher

COPY . .

RUN cp .netrc /root/.netrc

RUN dep ensure -vendor-only -v

RUN go build -o /usr/local/bin/webapp ./cmd/web/web.go

RUN rm -rf /root/.netrc

FROM alpine:3.8 AS app

RUN apk add --update ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

COPY --from=builder /usr/local/bin/webapp /usr/local/bin/webapp

CMD webapp start -p 8001

EXPOSE 8001
