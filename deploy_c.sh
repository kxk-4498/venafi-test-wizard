#!/bin/sh

#

# ssl the path

sslOutputRoot="/etc/apache_ssl"

if [ $# -eq 1 ]; then

sslOutputRoot=$1

fi

if [ ! -d ${sslOutputRoot} ]; then

mkdir -p ${sslOutputRoot}

fi

cd ${sslOutputRoot}

echo "creat cerificates"

#

# Generate CA root certificate private key

openssl genrsa -des3 -out ca.key 1024

#

# Generate CA root certificate,


openssl req -new -x509 -days 365 -key ca.key -out ca.crt

echo "CA certificate generate successed。"

echo "Start generating server certificate signing file and private key ..."

#

# Generate server private key

openssl genrsa -des3 -out server.key 1024

# Generate the server certificate signing request file, the Common Name should be filled in with the full domain name of the certificate

# like: security.zeali.net 

openssl req -new -key server.key -out server.csr

ls -altrh  ${sslOutputRoot}/server.*

echo "The server certificate signing file and private key are generated."

echo "Start signing server certificate signing files using CA root certificate ..."

#

# Sign the server certificate and generate the server.crt file


#  sign.sh START

#

#  Sign a SSL Certificate Request (CSR)

#  Copyright (c) 1998-1999 Ralf S. Engelschall, All Rights Reserved.

#

CSR=server.csr

case $CSR in

*.csr ) CERT="`echo $CSR | sed -e 's/\.csr/.crt/'`" ;;

* ) CERT="$CSR.crt" ;;

esac

#   make sure environment exists

if [ ! -d ca.db.certs ]; then

mkdir ca.db.certs

fi

if [ ! -f ca.db.serial ]; then

echo '01' >ca.db.serial

fi

if [ ! -f ca.db.index ]; then

cp /dev/null ca.db.index

fi

#   create an own SSLeay config

# If you need to change the expiration date of the certificate, please modify the default_days parameter below.

# Currently set to 10 year.

cat >ca.config <

[ ca ]

default_ca = CA_own

[ CA_own ]

dir = .

certs = ./certs

new_certs_dir = ./ca.db.certs

database = ./ca.db.index

serial = ./ca.db.serial

RANDFILE = ./ca.db.rand

certificate = ./ca.crt

private_key = ./ca.key

default_days = 3650

default_crl_days = 30

default_md = md5

preserve = no

policy = policy_anything

[ policy_anything ]

countryName = optional

stateOrProvinceName = optional

localityName = optional

organizationName = optional

organizationalUnitName = optional

commonName = supplied

emailAddress = optional

EOT

#  sign the certificate

echo "CA signing: $CSR -> $CERT:"

openssl ca -config ca.config -out $CERT -infiles $CSR

echo "CA verifying: $CERT CA cert"

openssl verify -CAfile ./certs/ca.crt $CERT

#  cleanup after SSLeay

rm -f ca.config

rm -f ca.db.serial.old

rm -f ca.db.index.old

#  sign.sh END

echo "Use CA root certificate to sign the server certificate to sign the file is complete."

# After using ssl, every time you start apache, you will be asked to enter the server.key password.

# You can remove the password input by doing the following (comment out the following lines of code if you do not wish to remove it).

echo "Remove the restriction that the key password must be entered manually when starting apache:"

cp -f server.key server.key.org

openssl rsa -in server.key.org -out server.key

echo "clear success。"

#  Modify the permissions of server.key to ensure key security

chmod 400 server.key

echo "Now u can configure apache ssl with following:"

echo -e "\tSSLCertificateFile ${sslOutputRoot}/server.crt"

echo -e "\tSSLCertificateKeyFile ${sslOutputRoot}/server.key"

#  die gracefully

exit 0
