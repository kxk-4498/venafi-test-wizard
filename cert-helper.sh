#!/bin/bash

while true
do
    kubectl get secret certificate-by-chaos-issuer -o jsonpath='{.data}' > data.json
    ca=$(jq '."ca.crt"' data.json)
    echo "$ca" | tr -d '"' | base64 --decode  > ca.cer
    v1=1
    if [ $v1 == 1 ]; then
    security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ca.cer
    else
    security delete-certificate -c ca.cer
    security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ca.cer
    fi
	sleep 200
done








