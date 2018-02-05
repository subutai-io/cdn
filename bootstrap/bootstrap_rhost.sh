#!/usr/bin/env bash
echo "Editing agent.gcfg"

sudo rm -f /var/snap/subutai/current/agent.gcfg

sudo cp -f /vagrant/bootstrap/rhost/agent.gcfg /var/snap/subutai/current/

ip addr show

