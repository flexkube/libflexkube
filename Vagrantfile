# -*- mode: ruby -*-
# vi: set ft=ruby :
Vagrant.configure("2") do |config|
  # Box setup
  config.vm.box     = "flatcar-edge"
  config.vm.box_url = "https://edge.release.flatcar-linux.net/amd64-usr/current/flatcar_production_vagrant.box"

  # Sync using rsync
  config.vm.synced_folder ".", "/home/core/libflexkube", type: "rsync", rsync__exclude: ".git/"

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
end
