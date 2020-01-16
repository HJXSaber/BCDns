#!/bin/bash

cd /go/src/

cat /tmp/hosts | while read line
do
    ip=$(echo $line | awk '{print $1}')
    expect -c "
        spawn scp ./update.sh root@$ip:/go/src/
	    expect {
            \"*yes/no*\" {send \"yes\r\";exp_continue;}
            \"*assword\" {set timeout 300; send \"123456\r\"; exp_continue;}
        }"
done