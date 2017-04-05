FROM golang:1.8-alpine
COPY . /go/src/app
RUN apk add --update git make && \
    cd /go/src/app/ && \
    make
VOLUME /config
CMD /go/src/app/bin/mysb -c /config/config.yaml
