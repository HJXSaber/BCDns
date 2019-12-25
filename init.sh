#!/bin/bash

mode="docker"


if [[ $1 == ${mode} ]]; then
    cd /var/run/bcdns
    for i in $(seq 1 $2)
    do
        for j in $(seq 1 $2)
        do
            if [[ ${i} -ne ${j} ]]; then
                cp "s$i/LocalCertificate.cer" "s$j/s$i.cer"
            fi
        done
    done
else
    if [ ! -d "./certificateAuthority/conf/$HOST" ]; then
        mkdir ./certificateAuthority/conf/$HOST
    fi
    expect -c "
    spawn scp onos@$1:/workspace/hosts /tmp
    expect {
        \"*assword\" {set timeout 300; send \"123456\r\"; exp_continue;}
        \"yes/no\" {send \"yes\r\";}
    }"
    cat /tmp/hosts | while read line
    do
        ip=$(echo $line | awk '{print $1}')
        hostname=$(echo $line | awk '{print $2}')
        if [[ $HOST == $hostname ]]; then
            continue
        fi
        expect -c "
        spawn scp ./certificateAuthority/conf/$HOST/LocalCertificate.cer onos@$ip:/go/src/BCDns_0.1/certificateAuthority/conf/$hostname/$HOST.cer
        expect {
        \"*yes/no*\" {send \"yes\r\";exp_continue;}
        \"*assword\" {set timeout 300; send \"123456\r\"; exp_continue;}
        }"
    done
fi