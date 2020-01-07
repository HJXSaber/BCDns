#!/bin/bash

cd /go/src/BCDns_0.1

cat /tmp/hosts | while read line
do
    ip=$(echo $line | awk '{print $1}')
    hostname=$(echo $line | awk '{print $2}')
    ./generateCertByIp.sh $ip CH BJ BJ BUPT 222 $hostname
    expect -c "
        spawn scp ./certificateAuthority/conf/tmp/$hostname.cer root@$ip:/go/src/BCDns_0.1/certificateAuthority/conf/$hostname/LocalCertificate.cer
	    expect {
            \"*yes/no*\" {send \"yes\r\";exp_continue;}
            \"*assword\" {set timeout 300; send \"123456\r\"; exp_continue;}
        }"
done