#!/bin/bash
while true
do
    kubectl get secret certificate-by-chaos-issuer -o jsonpath='{.data}' > data.json
    ca=$(jq '."ca.crt"' data.json)
    echo "$ca" | tr -d '"' | base64 --decode  > ca.crt
    v1=1
    if [ $v1 == 1 ]; then
    dir /usr/local/share/ca-certificates/
    sudo cp ca.crt /usr/local/share/ca-certificates/ca.crt
    sudo update-ca-certificates
    v1=2
    else
    sudo rm /usr/local/share/ca-certificates/ca.crt
    sudo update-ca-certificates --fresh
    sudo cp ca.crt /usr/local/share/ca-certificates/ca.crt
    sudo update-ca-certificates
    fi
    rm data.json
    rm ca.crt
    sleep 10
done
