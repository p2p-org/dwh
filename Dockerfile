FROM golang:1.14-alpine

RUN apk add --update --no-cache bash ca-certificates git libc-dev make build-base

#ENV PATH /go/bin:$PATH
#ENV GOPATH /go
ENV DWHPATH /go/src/github.com/corestario/dwh/
#RUN mkdir -p $DWHPATH
WORKDIR $DWHPATH

ENV GO111MODULE=off

COPY . $DWHPATH
COPY ./config.toml /root/config.toml

ARG APPNAME
ENV APP=$APPNAME
RUN go install $DWHPATH/cmd/$APP



EXPOSE 11535

ENTRYPOINT $APP
