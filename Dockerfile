FROM golang:1.13.0

WORKDIR $GOPATH/src

COPY ./BCDns_0.1 ./BCDns_0.1

COPY ./BCDns_client ./BCDns_client

ENV GO111MODULE="on" GOPROXY="https://goproxy.cn"

RUN apt update && apt install net-tools && cd BCDns_0.1 && go mod tidy