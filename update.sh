#!/bin/bash

cd /go/src/BCDns_0.1 && git stash

expect -c "
    spawn git pull origin serial-exec
    expect {
        \"Username*\" {set timeout 300; send \"HJXSaber\r\"; exp_continue;}
        \"Password*\" {set timeout 300; send \"Jiangxue104\r\"; exp_continue;}
    }"