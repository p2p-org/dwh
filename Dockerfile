FROM golang:1.14-alpine

RUN apk add --update --no-cache bash ca-certificates git libc-dev make build-base

ENV DWHPATH /go/src/github.com/p2p-org/dwh/
#WORKDIR $DWHPATH/..
#
#RUN git clone https://github.com/corestario/cosmos-utils
#RUN cd cosmos-utils && git checkout merge-v0.33.4
#
#RUN git clone https://github.com/corestario/modules
#RUN cd modules && git checkout hack-ibc
#
#RUN git clone https://github.com/p2p-org/marketplace
#RUN cd marketplace && git checkout feat/hack-ibc-nft

ENV GO111MODULE=on
WORKDIR $DWHPATH
COPY . .
ARG APPNAME
ENV APP=$APPNAME
RUN go install $DWHPATH/cmd/$APP

COPY ./config.toml /root/config.toml

EXPOSE 11535
ENTRYPOINT $APP
