#!/usr/bin/env bash

cd /var/run/bcdns

for i in $(seq 1 $1)
do
    for j in $(seq 1 $1)
    do
        if [[ ${i} -ne ${j} ]]; then
            cp "s$i/LocalCertificate.cer" "s$j/s$i.cer"
        fi
    done
done