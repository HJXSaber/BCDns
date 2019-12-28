#!/bin/bash
#
#export BCDNSConfFile="/var/opt/go/src/BCDns_0.1/bcDns/conf/s1/BCDNS"
#export CertificatesPath="/var/opt/go/src/BCDns_0.1/certificateAuthority/conf/s1/"
#
#cd bcDns/cmd
#
#rm -rf blockchain_*
#
#go run main.go

pid=$(ps -ef| grep -e "go run main.go"| grep -v "grep"| awk '{print $2}')

kill -9 $pid