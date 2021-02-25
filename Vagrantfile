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

  # Build route table for each node.
  #
  # TODO only required when kubenet network plugin is used.
  routes = []

  # This makes sure outgoing traffic to service CIDR use eth1 interface, so it gets correct source IP address.
  # Otherwise it use default eth0.
  r = <<EOF
[Route]
Destination=11.0.0.0/24
Scope=link
EOF
  routes.push(r)

  (2..253).each do |i|
    r = <<EOF
[Route]
Gateway=#{nodes_cidr}#{i}
Destination=10.1.#{i}.0/24
EOF
    routes.push(r)
  end

  # Also don't use Virtualbox DNS server, as it does not support TCP queries it seems and this makes CoreDNS complains.
  resolved_config = <<-EOF
    [Resolve]
    DNS=8.8.8.8 8.8.4.4
    Domains=~.
  EOF

  containerd_config = <<-EOF
# persistent data location
root = \\"/var/lib/containerd\\"
# runtime state information
state = \\"/run/docker/libcontainerd/containerd\\"
# set containerd as a subreaper on linux when it is not running as PID 1
subreaper = true
# set containerd's OOM score
oom_score = -999
# CRI plugin listens on a TCP port by default
disabled_plugins = []

# grpc configuration
[grpc]
address = \\"/run/docker/libcontainerd/docker-containerd.sock\\"
# socket uid
uid = 0
# socket gid
gid = 0

[plugins.linux]
# shim binary name/path
shim = \\"containerd-shim\\"
# runtime binary name/path
runtime = \\"runc\\"
# do not use a shim when starting containers, saves on memory but
# live restore is not supported
no_shim = false
# display shim logs in the containerd daemon's log output
shim_debug = true
  EOF

  containerd_override = <<-EOF
    [Service]
    Environment=CONTAINERD_CONFIG=/etc/containerd/config.toml
    ExecStart=
    ExecStart=/usr/bin/env PATH=\\${TORCX_BINDIR}:\\${PATH} \\${TORCX_BINDIR}/containerd --config \\${CONTAINERD_CONFIG}
  EOF

  common_provisioning_script = <<-EOF
    mkdir -p /etc/systemd/resolved.conf.d && echo "#{resolved_config}" | sudo tee /etc/systemd/resolved.conf.d/dns_servers.conf >/dev/null
    mkdir -p /etc/systemd/system/containerd.service.d && echo "#{containerd_override}" | sudo tee /etc/systemd/system/containerd.service.d/10-use-custom-config.conf >/dev/null
    mkdir -p /etc/containerd && echo "#{containerd_config}" | sudo tee /etc/containerd/config.toml >/dev/null
    sudo systemctl daemon-reload
    sudo systemctl enable iptables-store iptables-restore docker containerd systemd-timesyncd
    sudo systemctl stop update-engine locksmithd
    sudo systemctl mask update-engine locksmithd
    sudo systemctl start docker systemd-timesyncd iptables-store
    sudo systemctl restart systemd-networkd systemd-resolved containerd
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
        config.vm.synced_folder ".", "/home/core/libflexkube", type: "rsync", rsync__exclude: [".git/", "bin/", "e2e/config.yaml", "e2e/state.yaml", "e2e/resources", "e2e/values", "e2e/kubeconfig", "libvirt"]

        # Read content of Vagrant SSH private key.
        ssh_private_key = File.read(ENV['HOME'] + "/.vagrant.d/insecure_private_key")

        # Primary node provisioning.
        config.vm.provision "shell", inline: <<-EOF
          set -e
          echo "#{ssh_private_key}" > /home/core/.ssh/id_rsa && chmod 0600 /home/core/.ssh/id_rsa
          openssl rand -base64 14 > /home/core/.password
          yes $(cat /home/core/.password) | sudo passwd core
        EOF
      end

      # Controller provisioning.
      config.vm.provision "shell", inline: <<-EOF
        set -e
        mkdir -p /etc/systemd/network/50-vagrant1.network.d && echo "#{(routes - [routes[i]]).join("\n")}" | sudo tee /etc/systemd/network/50-vagrant1.network.d/routes.conf >/dev/null
        sudo iptables -t nat -A POSTROUTING -s 10.1.#{i+1}.0/24 -j SNAT --destination 10.0.2.0/24 --to-source 10.0.2.15
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
        mkdir -p /etc/systemd/network/50-vagrant1.network.d && echo "#{(routes - [routes[i+controllers]]).join("\n")}" | sudo tee /etc/systemd/network/50-vagrant1.network.d/routes.conf >/dev/null
        sudo iptables -t nat -A POSTROUTING -s 10.1.#{i+controllers+1}.0/24 -j SNAT --destination 10.0.2.0/24 --to-source 10.0.2.15
        #{common_provisioning_script}
      EOF
    end
  end
end
