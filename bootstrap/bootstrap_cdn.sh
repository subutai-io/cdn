#!/usr/bin/env bash
echo "Installing NGINX to cdn proxy ..."

sudo apt-get update
sudo apt-get install -y nginx

sudo mkdir -p /etc/ssl/certs/
sudo mkdir -p /etc/ssl/private/
sudo mkdir -p /var/snap/subutai/common/cache/nginx/templ/
sudo cp -f /vagrant/bootstrap/cdn/$1.conf /etc/nginx/sites-enabled/
sudo cp -f /vagrant/bootstrap/cdn/nginx-selfsigned.crt /etc/ssl/certs/
sudo cp -f /vagrant/bootstrap/cdn/nginx-selfsigned.key /etc/ssl/private/

sudo /etc/init.d/nginx restart
sudo /etc/init.d/nginx status
ip addr show

