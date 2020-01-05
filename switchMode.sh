#!/bin/bash

cd /go/src/BCDns_0.1

if [[ $1 -eq 1 ]]; then
    git checkout serial-exec
elif [[ $1 -eq 2 ]]; then
    git checkout pbft
fi