# -*- mode: ruby -*-
# vi: set ft=ruby :
Vagrant.configure("2") do |config|
  # Box setup
  config.vm.box     = "flatcar-edge"
  config.vm.box_url = "https://edge.release.flatcar-linux.net/amd64-usr/current/flatcar_production_vagrant.box"

  # Sync using rsync, but don't copy locally built binaries and don't remove Terraform files from virtual machine
  config.vm.synced_folder ".", "/home/core/libflexkube", type: "rsync", rsync__exclude: [".git/", "bin/", "e2e/terraform.tfstate*", "e2e/.terraform"]

  # Virtualbox + resources
  config.vm.provider :virtualbox do |v|
    v.check_guest_additions = false
    v.functional_vboxsf     = false
    v.cpus                  = 6
    v.memory                = 4096
    v.customize ['modifyvm', :id, '--paravirtprovider', 'kvm']
  end

  # SSH
  config.ssh.username   = 'core'

  # Provisioning
  config.vm.provision "shell", inline: <<-EOF
    set -e
    ssh-keygen -b 2048 -t rsa -f /home/core/.ssh/id_rsa -q -N ''
    cat /home/core/.ssh/id_rsa.pub >> /home/core/.ssh/authorized_keys.d/generated
    sudo update-ssh-keys
    openssl rand -base64 14 > /home/core/.password
    yes $(cat /home/core/.password) | sudo passwd core
    # SNAT traffic from pods (10.1.0.0/24) to e.g. DNS server (10.0.2.3 by default) using host IP (10.0.2.15)
    sudo iptables -t nat -A POSTROUTING -s 10.1.0.0/24 -j SNAT --destination 10.0.2.0/24 --to-source 10.0.2.15
    sudo systemctl enable iptables-store iptables-restore docker
    sudo systemctl start docker
  EOF

  # Forward kube-apiserver port to host
  config.vm.network "forwarded_port", guest: 6443, host: 6443
end
