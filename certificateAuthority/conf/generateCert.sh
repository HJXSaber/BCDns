#!/usr/bin/env bash

rm -f Local*

config="authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names

[alt_names]
IP.1="

ip=$(ifconfig |grep "ens33" -C 1|grep "inet"|awk  '{print $2}')

sed "s/0.0.0.0/$ip/g" v3.txt -i

openssl req -nodes -newkey rsa:1024 -keyout LocalPrivate.pem -out LocalCertificate.csr -subj "/C=CH/ST=BJ/L=BJ/O=BUPT/OU=222/CN=s1"

openssl x509 -req -days 365 -sha1 -extfile v3.txt -CA RootCertificate.cer -CAkey RootPrivate.pem -CAserial ca.srl -in LocalCertificate.csr -out LocalCertificate.cer