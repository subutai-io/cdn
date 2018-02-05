#!/usr/bin/env bash
echo "Installing NGINX to cdn proxy ..."

sudo apt-get update
sudo apt-get install -y nginx

sudo mkdir -p /etc/ssl/certs/
sudo mkdir -p /etc/ssl/private/
sudo mkdir -p /var/snap/subutai/common/cache/nginx/templ/
sudo cp -f /vagrant/bootstrap/cdn/$1.conf /etc/nginx/sites-enabled/

echo "

openssl rand -base64 48 > passphrase.txt

# Generate a Private Key
openssl genrsa -aes128 -passout file:passphrase.txt -out server.key 2048

# Generate a CSR (Certificate Signing Request)
openssl req -new -passin file:passphrase.txt -key server.key -out server.csr -subj "/C=FR/O=krkr/OU=Domain Control Validated/CN=*.krkr.io"

# Remove Passphrase from Key
cp server.key server.key.org
openssl rsa -in server.key.org -passin file:passphrase.txt -out server.key

# Generating a Self-Signed Certificate for 100 years
openssl x509 -req -days 36500 -in server.csr -signkey server.key -out server.crt

mv server.crt ssl.crt
mv server.key ssl.key" > cert.sh
chmod +x cert.sh
./cert.sh

sudo mv ssl.crt /etc/ssl/certs/ssl.crt
sudo mv ssl.key /etc/ssl/private/ssl.key

sudo /etc/init.d/nginx restart
sudo /etc/init.d/nginx status
ip addr show

