#!/bin/bash

cd /go/src/BCDns_0.1/messages

go test -v messages_test.go messages.go > log

grep count log| awk '{print $2}'