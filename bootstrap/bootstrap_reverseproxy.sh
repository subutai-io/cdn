#!/usr/bin/env bash
echo "Installing NGINX to reverse proxy ..."

sudo apt-get update
sudo apt-get install -y nginx

sudo mkdir -p /var/snap/subutai-dev/current/nginx-includes/http/
sudo chmod -R 777 /etc/nginx/nginx.conf

sudo mkdir -p /etc/nginx/
sudo touch /etc/nginx/proxy.conf
sudo chmod -R 777 /etc/nginx/proxy.conf

sudo mv /etc/nginx/nginx.conf /etc/nginx/nginx.conf.bak
sudo mv /etc/nginx/proxy.conf /etc/nginx/proxy.conf.bak
sudo mv /var/snap/subutai-dev/current/nginx-includes/http/gorjun.conf /var/snap/subutai-dev/current/nginx-includes/http/gorjun.conf.bak

sudo cp -f /vagrant/bootstrap/reverseproxy/nginx.conf /etc/nginx/
sudo cp -f /vagrant/bootstrap/reverseproxy/proxy.conf /etc/nginx/
sudo cp -f /vagrant/bootstrap/reverseproxy/gorjun.conf /var/snap/subutai-dev/current/nginx-includes/http/

sudo mkdir -p /etc/nginx/ssl
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

sudo mv ssl.crt /etc/nginx/ssl/ssl.crt
sudo mv ssl.key /etc/nginx/ssl/ssl.key


sudo /etc/init.d/nginx restart
sudo /etc/init.d/nginx status

ip addr show

