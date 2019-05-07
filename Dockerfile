FROM golang:1.11-alpine3.8

RUN apk --update upgrade && \
    apk add curl tzdata ca-certificates git make && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

RUN go version

RUN mkdir -p /gocode/

ENV GOPATH /gocode/
ENV CGO_ENABLED=0

RUN go get github.com/golang/mock/gomock

RUN go get -u golang.org/x/lint/golint

# Expose 8001 if running the webapp.
EXPOSE 8001
