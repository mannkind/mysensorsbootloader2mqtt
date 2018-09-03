FROM golang:1.11-alpine3.8 as build
COPY . /go/src/app
RUN apk add --no-cache --update build-base git && \
    cd /go/src/app/ && \
    make

FROM alpine:3.8
COPY --from=build /go/src/app/bin/mysb /usr/local/bin/mysb
VOLUME /config
CMD mysb -c /config/config.yaml
