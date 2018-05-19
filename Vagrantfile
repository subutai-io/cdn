# -*- mode: ruby -*-
# vi: set ft=ruby :

# Change this to wherever you are planning on storing your project
PROJECT_REPOSITORY='github.com/subutai-io/cdn'
BOX_IMAGE = "debian/stretch64"

NODE_COUNT = 1
MASTER_RAM = 800
MASTER_CPU = 2
CDN_RAM = 384
CDN_CPU = 1

Vagrant.configure(2) do |config|

  config.vm.define "master" do |subconfig|
    subconfig.vm.box = BOX_IMAGE
    subconfig.vm.hostname = "master"
    subconfig.vm.provider :virtualbox do |vb|
        vb.gui = true
    end
    subconfig.vm.network "private_network", ip: "10.55.2.10",
         virtualbox__intnet: true
    subconfig.vm.network "public_network"
    subconfig.vm.provider "virtualbox" do |v|
         v.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
         v.memory = MASTER_RAM
         v.cpus = MASTER_CPU
    end
    subconfig.vm.synced_folder ".", "/home/vagrant/go/src/#{PROJECT_REPOSITORY}", create: true, type: 'rsync'
    subconfig.vm.provision "shell", path: 'bootstrap/bootstrap.sh', privileged: false, env: {'PROJECT_REPOSITORY' => PROJECT_REPOSITORY}
    end

   config.vm.define "reverseproxy" do |subconfig|
      subconfig.vm.box = BOX_IMAGE
      subconfig.vm.hostname = "reverseproxy"
      subconfig.vm.network "public_network"
      subconfig.vm.network "private_network", ip: "10.55.2.100",
               virtualbox__intnet: true
      subconfig.vm.network "public_network"
      subconfig.vm.provider "virtualbox" do |v|
        v.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
        v.memory = CDN_RAM
        v.cpus = CDN_CPU
      end
      subconfig.vm.provision "shell", path: 'bootstrap/bootstrap_reverseproxy.sh', privileged: false
    end

    (1..NODE_COUNT).each do |i|
        config.vm.define "cdn#{i}" do |subconfig|
          subconfig.vm.box = BOX_IMAGE
          subconfig.vm.hostname = "cdn#{i}"
          subconfig.vm.network "public_network"
          subconfig.vm.network "private_network", ip: "10.55.2.#{i + 10}",
            virtualbox__intnet: true
          subconfig.vm.provider "virtualbox" do |v|
            v.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
            v.memory = CDN_RAM
            v.cpus = CDN_CPU
          end
          subconfig.vm.provision "shell", path: 'bootstrap/bootstrap_cdn.sh', privileged: false, :args => subconfig.vm.hostname
        end
      end


   config.vm.define "rhost" do |subconfig|
     subconfig.vm.box = "subutai/stretch"
     subconfig.vm.hostname = "rhost"
     subconfig.vm.network "public_network"
     subconfig.vm.provider "virtualbox" do |v|
       v.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
       v.memory = MASTER_RAM
       v.cpus = MASTER_CPU
     end
     subconfig.vm.provision "shell", path: 'bootstrap/bootstrap_rhost.sh', privileged: false

   end
end
