FROM alpine:3.8

RUN apk add --update ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

COPY ./bin/webapp /usr/local/bin/webapp

CMD webapp start -p 8001

EXPOSE 8001
