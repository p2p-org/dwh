FROM golang:1.14-alpine

RUN apk add --update --no-cache bash ca-certificates git libc-dev make build-base

#ENV PATH /go/bin:$PATH
#ENV GOPATH /go
ENV DWHPATH /go/src/github.com/p2p-org/dwh/
#RUN mkdir -p $DWHPATH
WORKDIR $DWHPATH/..

RUN git clone https://github.com/corestario/cosmos-utils
RUN cd cosmos-utils && git checkout merge-v0.33.4

RUN git clone https://github.com/corestario/modules
RUN cd modules && git checkout hack-ibc

RUN git clone https://github.com/p2p-org/marketplace
RUN cd marketplace && git checkout feat/hack-ibc-nft

#ENV GO111MODULE=off

COPY . $DWHPATH
COPY ./config.toml /root/config.toml

ARG APPNAME
ENV APP=$APPNAME

WORKDIR $DWHPATH

RUN go install $DWHPATH/cmd/$APP

EXPOSE 11535

ENTRYPOINT $APP
