# -*- mode: ruby -*-
# vi: set ft=ruby :
Vagrant.configure("2") do |config|
  channel = ENV["TF_VAR_flatcar_channel"] || "edge"

  # Box setup.
  config.vm.box     = "flatcar-#{channel}"
  config.vm.box_url = "https://#{channel}.release.flatcar-linux.net/amd64-usr/current/flatcar_production_vagrant.box"

  # Virtualbox resources.
  config.vm.provider :virtualbox do |v|
    v.check_guest_additions = false
    v.functional_vboxsf     = false
    v.cpus                  = 2
    v.memory                = 2048
    v.customize ['modifyvm', :id, '--paravirtprovider', 'kvm']
  end

  # SSH.
  config.ssh.username = 'core'

  # Number of controllers and workers.
  controllers = ENV["TF_VAR_controllers_count"].to_i || 1
  workers = ENV["TF_VAR_workers_count"].to_i || 0
  nodes_cidr = (ENV["TF_VAR_nodes_cidr"] || "192.168.50.0/24").split("/")[0].split(".")[0..2].join(".") + "."

  # Make sure there is only one primary VM.
  primary = true

  # Also don't use Virtualbox DNS server, as it does not support TCP queries it seems and this makes CoreDNS complains.
  network_config = <<-EOF
[Match]
Name=eth0

[Network]
DHCP=yes
DNS=8.8.8.8
DNS=8.8.4.4

[DHCP]
UseDNS=false
EOF

  common_provisioning_script = <<-EOF
    mkdir -p /etc/systemd/network && echo "#{network_config}" | sudo tee /etc/systemd/network/10-virtualbox.network >/dev/null
    sudo systemctl daemon-reload
    sudo systemctl enable iptables-store iptables-restore docker containerd systemd-timesyncd
    sudo systemctl stop update-engine locksmithd
    sudo systemctl mask update-engine locksmithd
    sudo systemctl start docker systemd-timesyncd iptables-store
    sudo systemctl restart systemd-networkd
  EOF

  # Controllers.
  (1..controllers).each do |i|
    config.vm.define vm_name = "controller%02d" % i, primary: primary do |config|
      # Set hostname
      config.vm.hostname = vm_name

      config.vm.network "private_network", ip: nodes_cidr + (i+1).to_s

      if primary
        primary = false

        config.vm.provider :virtualbox do |v|
          v.cpus                  = 6
          v.memory                = 4096
        end

        # Sync using rsync, but don't copy locally built binaries and don't remove Terraform files from virtual machine.
        config.vm.synced_folder ".", "/home/core/libflexkube", type: "rsync", rsync__exclude: [
          ".git/",
          "bin/",
          "e2e/config.yaml",
          "e2e/test-config.yaml",
          "e2e/state.yaml",
          "e2e/resources",
          "e2e/values",
          "e2e/kubeconfig",
          "libvirt",
        ]

        # Read content of Vagrant SSH private key.
        ssh_private_key = File.read(ENV['HOME'] + "/.vagrant.d/insecure_private_key")

        # Primary node provisioning.
        config.vm.provision "shell", inline: <<-EOF
          set -e
          echo "#{ssh_private_key}" > /home/core/.ssh/id_rsa && chmod 0600 /home/core/.ssh/id_rsa
          openssl rand -base64 14 > /home/core/.ssh/password
          yes $(cat /home/core/.ssh/password) | sudo passwd core
        EOF
      end

      # Controller provisioning.
      config.vm.provision "shell", inline: <<-EOF
        set -e
        #{common_provisioning_script}
      EOF
    end
  end

  # Workers.
  (1..workers).each do |i|
    config.vm.define vm_name = "worker%02d" % i do |config|
      config.vm.hostname = vm_name
      config.vm.network "private_network", ip: nodes_cidr + (i+1+10).to_s

      # Provisioning.
      config.vm.provision "shell", inline: <<-EOF
        set -e
        #{common_provisioning_script}
      EOF
    end
  end
end
