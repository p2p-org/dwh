FROM golang:1.12-alpine3.10

ARG APPNAME
ENV APP=$APPNAME
RUN apk update
RUN apk upgrade
RUN apk add bash ca-certificates git libc-dev

ENV GO111MODULE=off
ENV PATH /go/bin:$PATH
ENV GOPATH /go
ENV DWHPATH /go/src/github.com/dgamingfoundation/dwh/
RUN mkdir -p $DWHPATH

COPY . $DWHPATH
COPY ./config.toml /root/config.toml

RUN go install $DWHPATH/cmd/$APP

WORKDIR /root/

EXPOSE 11535

ENTRYPOINT $APP
