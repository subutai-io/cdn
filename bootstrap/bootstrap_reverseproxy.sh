#!/usr/bin/env bash
echo "Installing NGINX to reverse proxy ..."

sudo apt-get update
sudo apt-get install -y nginx

sudo mkdir -p /var/snap/subutai-dev/current/nginx-includes/http/
sudo chmod -R 777 /etc/nginx/nginx.conf

sudo mkdir -p /etc/nginx/
sudo touch /etc/nginx/proxy.conf
sudo chmod -R 777 /etc/nginx/proxy.conf

sudo rm -f /etc/nginx/nginx.conf
sudo rm -f /etc/nginx/proxy.conf
sudo rm -f /var/snap/subutai-dev/current/nginx-includes/http/gorjun.conf

sudo cp -f /vagrant/bootstrap/reverseproxy/nginx.conf /etc/nginx/
sudo cp -f /vagrant/bootstrap/reverseproxy/proxy.conf /etc/nginx/
sudo cp -f /vagrant/bootstrap/reverseproxy/gorjun.conf /var/snap/subutai-dev/current/nginx-includes/http/

sudo mkdir -p /etc/nginx/ssl
sudo cp -f /vagrant/bootstrap/reverseproxy/nginx.crt /etc/nginx/ssl/
sudo cp -f /vagrant/bootstrap/reverseproxy/nginx.key /etc/nginx/ssl/

sudo /etc/init.d/nginx restart
sudo /etc/init.d/nginx status

ip addr show

