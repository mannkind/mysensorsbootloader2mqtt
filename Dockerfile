FROM golang:1.8-alpine
COPY . /go/src/app
RUN apk add --no-cache --virtual .build-deps git make && \
    cd /go/src/app/ && \
    make && \
    apk del .build-deps && \
    mv /go/src/app/bin/mysb /go/bin && \
    rm -rf /go/src/app/*

VOLUME /config
CMD mysb -c /config/config.yaml
