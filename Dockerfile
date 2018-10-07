FROM golang:1.10-alpine3.7

RUN apk --update upgrade && \
    apk add curl tzdata ca-certificates git make && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

RUN go version

RUN curl https://raw.githubusercontent.com/golang/dep/v0.5.0/install.sh | sh

RUN dep version

RUN mkdir -p /gocode/

ENV GOPATH /gocode/

RUN go get github.com/golang/mock/gomock

RUN go install github.com/golang/mock/mockgen

# Expose 8001 if running the webapp.
EXPOSE 8001
