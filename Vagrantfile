BOX_IMAGE = "debian/stretch64"
NODE_COUNT = 2
MASTER_RAM = 768
MASTER_CPU = 2
CDN_RAM = 384
CDN_CPU = 1

Vagrant.configure("2") do |config|
  config.vm.define "master" do |subconfig|
    subconfig.vm.box = BOX_IMAGE
    subconfig.vm.hostname = "master"

    subconfig.vm.network "private_network", ip: "10.55.2.10",
      virtualbox__intnet: true

    subconfig.vm.provider "virtualbox" do |v|
      v.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
      v.memory = MASTER_RAM
      v.cpus = MASTER_CPU
    end

    subconfig.vm.provision "shell", inline: <<-SHELL
      echo "Installing golang, building Gorjun, and running ..."
      apt-get install -y curl git
      curl -O https://storage.googleapis.com/golang/go1.8.4.linux-amd64.tar.gz
      tar -zxvf ./go1.8.4.linux-amd64.tar.gz
      chown -R root:root ./go
      mv go /usr/local
      rm -rf ./go1.8.4.linux-amd64.tar.gz
      echo 'export GOPATH=$HOME/work' >> ~vagrant/.profile
      echo 'export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin' >> ~vagrant/.profile
      su - vagrant -c 'mkdir /home/vagrant/work'
      su - vagrant -c 'go get github.com/subutai-io/gorjun'
      echo 'TODO: still need to configure and start up gorjun'
      echo 'TODO: lets also load some stuff into it to test'
    SHELL
  end

  (1..NODE_COUNT).each do |i|
    config.vm.define "cdn#{i}" do |subconfig|
      subconfig.vm.box = BOX_IMAGE
      subconfig.vm.hostname = "cdn#{i}"
      subconfig.vm.network "private_network", ip: "10.55.2.#{i + 10}",
        virtualbox__intnet: true

      subconfig.vm.provider "virtualbox" do |v|
        v.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
        v.memory = CDN_RAM
        v.cpus = CDN_CPU
      end

      subconfig.vm.provision "shell", inline: <<-SHELL
        echo "Installing NGINX cdn node ..."
        apt-get install -y nginx
        echo 'TODO: still need to configure nginx and cache'
      SHELL
    end
  end

  config.vm.provision "shell", inline: <<-SHELL
    apt-get install -y avahi-daemon libnss-mdns net-tools
  SHELL
end
