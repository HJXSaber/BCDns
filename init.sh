#!/bin/bash

export HOST=$(ifconfig |grep eth0 -C 1|grep "inet"|awk  '{print $2}'|awk -F "." '{print $4}')
export BCDNSConfFile="/go/src/BCDns_0.1/bcDns/conf/$HOST/BCDNS"
export CertificatesPath="/go/src/BCDns_0.1/certificateAuthority/conf/$HOST/"
bash /go/src/BCDns_0.1/init.sh

cd /go/src/BCDns_0.1/certificateAuthority/conf/
if [ ! -d "./$HOST" ]; then
    mkdir ./$HOST
fi

expect -c "
    spawn ./generateCert.sh eth0 CH BJ BJ BUPT 222 $HOST
    expect {
        \"*pass*\" {set timeout 300; send \"0401\r\"; exp_continue;}
    }"