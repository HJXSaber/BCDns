#!/bin/bash

cd /go/src/BCDns_0.1/bcDns/cmd

rm -rf blockchain_*

if [[ $# == 1 ]]; then
    sed "s/\(false\|true\)/$1/g" ../conf/$HOST/BCDNS.json -i
fi
go run main.go > run.log