FROM golang:1.12-alpine3.10

RUN apk update
RUN apk upgrade
RUN apk add bash ca-certificates git libc-dev

ENV GO111MODULE=off
ENV PATH /go/bin:$PATH
ENV GOPATH /go
ENV DWHPATH /go/src/github.com/dgamingfoundation/dwh/
RUN mkdir -p $DWHPATH

COPY . $DWHPATH
COPY ./cmd/imgworker/config.toml /root/config.toml

RUN go install $DWHPATH/cmd/imgworker

WORKDIR /root/

ENTRYPOINT imgworker
