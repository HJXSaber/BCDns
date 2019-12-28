#!/bin/bash

cd /go/src/BCDns_0.1/bcDns/cmd

rm -rf blockchain_*

go run main.go > run.log