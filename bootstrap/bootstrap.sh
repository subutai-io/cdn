#!/usr/bin/env bash
echo "Installing golang, building Gorjun, and running ..."

sudo chown -R vagrant:vagrant $HOME/go	# Work around bug in Vagrant
sudo apt-get update
sudo apt-get install -y git
if [ ! -d /usr/local/go ]; then
	wget -q "https://storage.googleapis.com/golang/go1.9.linux-amd64.tar.gz"
	sudo tar -C /usr/local -xvf go1.9.linux-amd64.tar.gz
	sudo rm go1.9.linux-amd64.tar.gz
fi
touch $HOME/.profile
echo "export PATH=\$PATH:/usr/local/go/bin" >> $HOME/.profile
echo "export GOPATH=/home/vagrant/go" >> $HOME/.profile
source $HOME/.profile
cd /home/vagrant/go/src/$PROJECT_REPOSITORY
go get

sudo chmod -R 777 /opt
mkdir /opt/gorjun

sudo touch /etc/systemd/system/gorjun.service
sudo chmod 777 /etc/systemd/system/gorjun.service

sudo cp -f /vagrant/bootstrap/gorjun.service /etc/systemd/system/gorjun.service


sudo systemctl daemon-reload
sudo systemctl start gorjun.service
sudo systemctl status gorjun.service

ip addr show