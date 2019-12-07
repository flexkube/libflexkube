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
  EOF
end
