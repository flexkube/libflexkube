# We must know controller node public keys and IP addresses in Terraform, they will be used for provisioning workers
resource "wireguard_peer" "controller_nodes" {
  count = var.controller_nodes_count
  keepers = {
    server_id = hcloud_server.controller_nodes[count.index].id
  }
}

# SSH Keys
resource "tls_private_key" "terraform" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "hcloud_ssh_key" "management" {
  count      = length(var.management_ssh_keys)
  name       = element(split(" ", element(var.management_ssh_keys, count.index)), 2)
  public_key = element(var.management_ssh_keys, count.index)
}

resource "hcloud_ssh_key" "terraform" {
  name       = "terraform"
  public_key = tls_private_key.terraform.public_key_openssh
}

# Names formatting
resource "null_resource" "controller_nodes_names_short" {
  count = var.controller_nodes_count

  triggers = {
    name = format("controller-%s-%d", var.environment, count.index + 1)
  }
}

resource "null_resource" "controller_nodes_names" {
  count = var.controller_nodes_count

  triggers = {
    name = format(
      "%s.%s",
      element(
        null_resource.controller_nodes_names_short.*.triggers.name,
        count.index,
      ),
      var.domain,
    )
  }
}

resource "hcloud_network" "network" {
  name     = "network"
  ip_range = var.hetzner_cidr
}

resource "hcloud_network_subnet" "network" {
  network_id   = hcloud_network.network.id
  type         = "server"
  network_zone = "eu-central"
  ip_range     = var.hetzner_cidr
}

resource "hcloud_server_network" "network" {
  count      = var.controller_nodes_count
  server_id  = hcloud_server.controller_nodes[count.index].id
  network_id = hcloud_network.network.id
  ip         = cidrhost(var.hetzner_cidr, count.index + 2)
}

# Servers creation
resource "hcloud_server" "controller_nodes" {
  count = var.controller_nodes_count
  name = element(
    null_resource.controller_nodes_names.*.triggers.name,
    count.index,
  )
  image       = var.hcloud_server_image
  server_type = var.controller_nodes_type
  location    = var.hcloud_server_location
  rescue      = "linux64"
  ssh_keys    = concat(hcloud_ssh_key.management.*.name, hcloud_ssh_key.terraform.*.name)
}

# Calculate wireguard properties, for each node (think like map function)
# Allows to use formatlist() on all CIDRs
resource "null_resource" "controller_wireguard_configuration" {
  count = var.controller_nodes_count
  triggers = {
    wireguard_ip   = cidrhost(var.wireguard_cidr, count.index + 1)
    wireguard_cidr = "${cidrhost(var.wireguard_cidr, count.index + 1)}/32"
    pods_cidr      = "${cidrhost(var.pods_cidr, count.index * 256)}/24"
  }
}

data "ct_config" "config" {
  count = var.controller_nodes_count
  content = <<EOF
networkd:
  units:
    - name: 30-wg0.network
      contents: |
        [Match]
        Name = wg0

        [Network]
        Address = ${element(
  null_resource.controller_wireguard_configuration.*.triggers.wireguard_cidr,
  count.index,
  )}
storage:
  files:
    - path: /etc/crio/crio.conf
      filesystem: root
      mode: 0600
      contents:
        inline: |

          # The CRI-O configuration file specifies all of the available configuration
          # options and command-line flags for the crio(8) OCI Kubernetes Container Runtime
          # daemon, but in a TOML format that can be more easily modified and versioned.
          #
          # Please refer to crio.conf(5) for details of all configuration options.

          # CRI-O reads its storage defaults from the containers-storage.conf(5) file
          # located at /etc/containers/storage.conf. Modify this storage configuration if
          # you want to change the system's defaults. If you want to modify storage just
          # for CRI-O, you can change the storage configuration options here.
          [crio]

          # Path to the "root directory". CRI-O stores all of its data, including
          # containers images, in this directory.
          #root = "/var/lib/containers/storage"

          # Path to the "run directory". CRI-O stores all of its state in this directory.
          #runroot = "/var/run/containers/storage"

          # Storage driver used to manage the storage of images and containers. Please
          # refer to containers-storage.conf(5) to see all available storage drivers.
          #storage_driver = ""

          # List to pass options to the storage driver. Please refer to
          # containers-storage.conf(5) to see all available storage options.
          #storage_option = [
          #]

          # If set to false, in-memory locking will be used instead of file-based locking.
          file_locking = true

          # Path to the lock file.
          file_locking_path = "/run/crio.lock"


          # The crio.api table contains settings for the kubelet/gRPC interface.
          [crio.api]

          # Path to AF_LOCAL socket on which CRI-O will listen.
          listen = "/var/run/crio/crio.sock"

          # IP address on which the stream server will listen.
          stream_address = "127.0.0.1"

          # The port on which the stream server will listen.
          stream_port = "0"

          # Enable encrypted TLS transport of the stream server.
          stream_enable_tls = false

          # Path to the x509 certificate file used to serve the encrypted stream. This
          # file can change, and CRI-O will automatically pick up the changes within 5
          # minutes.
          stream_tls_cert = ""

          # Path to the key file used to serve the encrypted stream. This file can
          # change, and CRI-O will automatically pick up the changes within 5 minutes.
          stream_tls_key = ""

          # Path to the x509 CA(s) file used to verify and authenticate client
          # communication with the encrypted stream. This file can change, and CRI-O will
          # automatically pick up the changes within 5 minutes.
          stream_tls_ca = ""

          # Maximum grpc send message size in bytes. If not set or <=0, then CRI-O will default to 16 * 1024 * 1024.
          grpc_max_send_msg_size = 16777216

          # Maximum grpc receive message size. If not set or <= 0, then CRI-O will default to 16 * 1024 * 1024.
          grpc_max_recv_msg_size = 16777216

          # The crio.runtime table contains settings pertaining to the OCI runtime used
          # and options for how to set up and manage the OCI runtime.
          [crio.runtime]

          # A list of ulimits to be set in containers by default, specified as
          # "<ulimit name>=<soft limit>:<hard limit>", for example:
          # "nofile=1024:2048"
          # If nothing is set here, settings will be inherited from the CRI-O daemon
          #default_ulimits = [
          #]

          # default_runtime is the _name_ of the OCI runtime to be used as the default.
          # The name is matched against the runtimes map below.
          default_runtime = "runc"

          # If true, the runtime will not use pivot_root, but instead use MS_MOVE.
          no_pivot = false

          # Path to the conmon binary, used for monitoring the OCI runtime.
          conmon = "/usr/libexec/crio/conmon"

          # Environment variable list for the conmon process, used for passing necessary
          # environment variables to conmon or the runtime.
          conmon_env = [
            "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
          ]

          # If true, SELinux will be used for pod separation on the host.
          selinux = false

          # Path to the seccomp.json profile which is used as the default seccomp profile
          # for the runtime.
          seccomp_profile = "/etc/crio/seccomp.json"

          # Used to change the name of the default AppArmor profile of CRI-O. The default
          # profile name is "crio-default-" followed by the version string of CRI-O.
          apparmor_profile = "crio-default"

          # Cgroup management implementation used for the runtime.
          cgroup_manager = "cgroupfs"

          # List of default capabilities for containers. If it is empty or commented out,
          # only the capabilities defined in the containers json file by the user/kube
          # will be added.
          default_capabilities = [
            "CHOWN", 
            "DAC_OVERRIDE", 
            "FSETID", 
            "FOWNER", 
            "NET_RAW", 
            "SETGID", 
            "SETUID", 
            "SETPCAP", 
            "NET_BIND_SERVICE", 
            "SYS_CHROOT", 
            "KILL", 
          ]

          # List of default sysctls. If it is empty or commented out, only the sysctls
          # defined in the container json file by the user/kube will be added.
          default_sysctls = [
          ]

          # List of additional devices. specified as
          # "<device-on-host>:<device-on-container>:<permissions>", for example: "--device=/dev/sdc:/dev/xvdc:rwm".
          #If it is empty or commented out, only the devices
          # defined in the container json file by the user/kube will be added.
          additional_devices = [
          ]

          # Path to OCI hooks directories for automatically executed hooks.
          hooks_dir = [
          ]

          # List of default mounts for each container. **Deprecated:** this option will
          # be removed in future versions in favor of default_mounts_file.
          default_mounts = [
          ]

          # Path to the file specifying the defaults mounts for each container. The
          # format of the config is /SRC:/DST, one mount per line. Notice that CRI-O reads
          # its default mounts from the following two files:
          #
          #   1) /etc/containers/mounts.conf (i.e., default_mounts_file): This is the
          #      override file, where users can either add in their own default mounts, or
          #      override the default mounts shipped with the package.
          #
          #   2) /usr/share/containers/mounts.conf: This is the default file read for
          #      mounts. If you want CRI-O to read from a different, specific mounts file,
          #      you can change the default_mounts_file. Note, if this is done, CRI-O will
          #      only add mounts it finds in this file.
          #
          #default_mounts_file = ""

          # Maximum number of processes allowed in a container.
          pids_limit = 1024

          # Maximum sized allowed for the container log file. Negative numbers indicate
          # that no size limit is imposed. If it is positive, it must be >= 8192 to
          # match/exceed conmon's read buffer. The file is truncated and re-opened so the
          # limit is never exceeded.
          log_size_max = -1

          # Whether container output should be logged to journald in addition to the kuberentes log file
          log_to_journald = false

          # Path to directory in which container exit files are written to by conmon.
          container_exits_dir = "/var/run/crio/exits"

          # Path to directory for container attach sockets.
          container_attach_socket_dir = "/var/run/crio"

          # If set to true, all containers will run in read-only mode.
          read_only = false

          # Changes the verbosity of the logs based on the level it is set to. Options
          # are fatal, panic, error, warn, info, and debug.
          log_level = "error"

          # The UID mappings for the user namespace of each container. A range is
          # specified in the form containerUID:HostUID:Size. Multiple ranges must be
          # separated by comma.
          uid_mappings = ""

          # The GID mappings for the user namespace of each container. A range is
          # specified in the form containerGID:HostGID:Size. Multiple ranges must be
          # separated by comma.
          gid_mappings = ""

          # The minimal amount of time in seconds to wait before issuing a timeout
          # regarding the proper termination of the container.
          ctr_stop_timeout = 0

            # The "crio.runtime.runtimes" table defines a list of OCI compatible runtimes.
            # The runtime to use is picked based on the runtime_handler provided by the CRI.
            # If no runtime_handler is provided, the runtime will be picked based on the level
            # of trust of the workload.
            
            [crio.runtime.runtimes.runc]
            runtime_path = "/usr/bin/runc"
            runtime_type = "oci"
            


          # The crio.image table contains settings pertaining to the management of OCI images.
          #
          # CRI-O reads its configured registries defaults from the system wide
          # containers-registries.conf(5) located in /etc/containers/registries.conf. If
          # you want to modify just CRI-O, you can change the registries configuration in
          # this file. Otherwise, leave insecure_registries and registries commented out to
          # use the system's defaults from /etc/containers/registries.conf.
          [crio.image]

          # Default transport for pulling images from a remote container storage.
          default_transport = "docker://"

          # The image used to instantiate infra containers.
          pause_image = "k8s.gcr.io/pause:3.1"

          # If not empty, the path to a docker/config.json-like file containing credentials
          # necessary for pulling the image specified by pause_image above.
          pause_image_auth_file = ""

          # The command to run to have a container stay in the paused state.
          pause_command = "/pause"

          # Path to the file which decides what sort of policy we use when deciding
          # whether or not to trust an image that we've pulled. It is not recommended that
          # this option be used, as the default behavior of using the system-wide default
          # policy (i.e., /etc/containers/policy.json) is most often preferred. Please
          # refer to containers-policy.json(5) for more details.
          signature_policy = ""

          # Controls how image volumes are handled. The valid values are mkdir, bind and
          # ignore; the latter will ignore volumes entirely.
          image_volumes = "mkdir"

          # List of registries to be used when pulling an unqualified image (e.g.,
          # "alpine:latest"). By default, registries is set to "docker.io" for
          # compatibility reasons. Depending on your workload and usecase you may add more
          # registries (e.g., "quay.io", "registry.fedoraproject.org",
          # "registry.opensuse.org", etc.).
          #registries = [
          # ]


          # The crio.network table containers settings pertaining to the management of
          # CNI plugins.
          [crio.network]

          # Path to the directory where CNI configuration files are located.
          network_dir = "/etc/cni/net.d/"

          # Paths to directories where CNI plugin binaries are located.
          plugin_dirs = [
            "/opt/cni/bin/",
          ]
    - path: /etc/cni/net.d/.keep
      filesystem: root
      mode: 0600
    - path: /etc/systemd/system.conf.d/kubernetes-accounting.conf
      filesystem: root
      mode: 0600
      contents:
        inline: |
          # Required for kubelet
          [Manager]
          DefaultCPUAccounting=yes
          DefaultMemoryAccounting=yes
    - path: /etc/wireguard/wg0.conf
      filesystem: root
      mode: 0600
      contents:
        inline: |
          [Interface]
          PrivateKey = ${element(wireguard_peer.controller_nodes.*.private, count.index)}
          ListenPort = ${var.wireguard_port}
          Address = ${element(
  null_resource.controller_wireguard_configuration.*.triggers.wireguard_cidr,
  count.index,
  )}
          SaveConfig = true
${replace(
  join(
    "\n",
    formatlist(
      "          [Peer]\n          PublicKey = %s\n          Endpoint = %s:${var.wireguard_port}\n          AllowedIPs = %s, %s",
      wireguard_peer.controller_nodes.*.public,
      hcloud_server_network.network.*.ip,
      null_resource.controller_wireguard_configuration.*.triggers.wireguard_cidr,
      null_resource.controller_wireguard_configuration.*.triggers.pods_cidr,
    ),
  ),
  "          [Peer]\n          PublicKey = ${element(wireguard_peer.controller_nodes.*.public, count.index)}\n          Endpoint = ${element(hcloud_server_network.network.*.ip, count.index)}:${var.wireguard_port}\n          AllowedIPs = ${element(
    null_resource.controller_wireguard_configuration.*.triggers.wireguard_cidr,
    count.index,
    )}, ${element(
    null_resource.controller_wireguard_configuration.*.triggers.pods_cidr,
    count.index,
  )}",
  "",
  )}
    - path: /var/lib/iptables/rules-save
      filesystem: root
      mode: 0600
      contents:
        inline: |
          *filter
          :INPUT ACCEPT [0:0]
          :FORWARD ACCEPT [0:0]
          :OUTPUT ACCEPT [0:0]
          -A INPUT -p icmp -m comment --comment "000 accept all icmp" -j ACCEPT
          -A INPUT -i lo -m comment --comment "001 accept all to lo interface" -j ACCEPT
          -A INPUT -i wg0 -m comment --comment "001 accept all to wg0 interface" -j ACCEPT
          -A INPUT -d 127.0.0.0/8 ! -i lo -m comment --comment "002 reject local traffic not on loopback interface" -j REJECT --reject-with icmp-port-unreachable
          -A INPUT -m comment --comment "003 accept related established rules" -m state --state RELATED,ESTABLISHED -j ACCEPT
${join(
  "\n",
  formatlist(
    "          -A INPUT -s %s -m comment --comment \"004 accept WireGuard traffic from peers\" -i eth1 -p udp -m multiport --dports ${var.wireguard_port} -j ACCEPT",
    hcloud_server_network.network.*.ip,
  ),
  )}
${join(
  "\n",
  formatlist(
    "          -A INPUT -s %s -i eth0 -p tcp -m multiport --dports ${var.ssh_port} -m comment --comment \"005 allow SSH from %s management CIDR\" -j ACCEPT",
    var.management_cidrs,
    var.management_cidrs,
  ),
  )}
          -A INPUT -m comment --comment "999 drop all" -j DROP
          COMMIT
    - path: /etc/sysctl.d/rancher-hardening.conf
      filesystem: root
      mode: 0644
      contents:
        inline: |
          vm.overcommit_memory=1
          kernel.panic=10
          kernel.panic_on_oops=1
    - path: /etc/containerd/config.toml
      filesystem: root
      mode: 0600
      contents:
        inline: |
          # persistent data location
          root = "/var/lib/containerd"
          # runtime state information
          state = "/run/docker/libcontainerd/containerd"
          # set containerd as a subreaper on linux when it is not running as PID 1
          subreaper = true
          # set containerd's OOM score
          oom_score = -999
          # CRI plugin listens on a TCP port by default
          disabled_plugins = []

          # grpc configuration
          [grpc]
          address = "/run/docker/libcontainerd/docker-containerd.sock"
          # socket uid
          uid = 0
          # socket gid
          gid = 0

          [plugins.linux]
          # shim binary name/path
          shim = "containerd-shim"
          # runtime binary name/path
          runtime = "runc"
          # do not use a shim when starting containers, saves on memory but
          # live restore is not supported
          no_shim = false
          # display shim logs in the containerd daemon's log output
          shim_debug = true

          [plugins.cri]
          stream_server_address = "127.0.0.1"
          stream_server_port = "10010"
          systemd_cgroup = true
    - path: /etc/ssh/sshd_config
      filesystem: root
      mode: 0600
      contents:
        inline: |
          # Use most defaults for sshd configuration.
          UsePrivilegeSeparation sandbox
          Subsystem sftp internal-sftp
          ClientAliveInterval 180
          UseDNS no
          UsePAM yes
          PrintLastLog no # handled by PAM
          PrintMotd no # handled by PAM

          PermitRootLogin yes
          AllowUsers core
          AllowUsers root
          AuthenticationMethods publickey
    - path: /etc/hostname
      filesystem: root
      mode: 0420
      contents:
        inline: |
          ${element(
  null_resource.controller_nodes_names_short.*.triggers.name,
  count.index,
)}
systemd:
  units:
    - name: locksmithd.service
      mask: true
    - name: docker.service
      enabled: true
    - name: iptables-restore.service
      enabled: true
    - name: crio.service
      enabled: true
    - name: wg-quick.service
      enabled: true
      contents: |
        [Unit]
        Description=WireGuard via wg-quick(8) for wg0
        After=network-online.target nss-lookup.target
        Wants=network-online.target nss-lookup.target
        Documentation=man:wg-quick(8)
        Documentation=man:wg(8)
        Documentation=https://www.wireguard.com/
        Documentation=https://www.wireguard.com/quickstart/
        Documentation=https://git.zx2c4.com/WireGuard/about/src/tools/man/wg-quick.8
        Documentation=https://git.zx2c4.com/WireGuard/about/src/tools/man/wg.8

        [Service]
        Type=oneshot
        RemainAfterExit=yes
        ExecStart=/usr/bin/wg-quick up wg0
        ExecStop=/usr/bin/wg-quick down wg0

        [Install]
        WantedBy=multi-user.target
    - name: containerd.service
      dropins:
      - name: 10-use-custom-config.conf
        contents: |
          [Service]
          Environment=CONTAINERD_CONFIG=/etc/containerd/config.toml
          ExecStart=
          ExecStart=/usr/bin/env PATH=$${TORCX_BINDIR}:$${PATH} $${TORCX_BINDIR}/containerd --config $${CONTAINERD_CONFIG}
    - name: sshd.socket
      dropins:
      - name: 10-sshd-port.conf
        contents: |
          [Socket]
          ListenStream=
          ListenStream=${var.ssh_port}
passwd:
  users:
    - name: core
      password_hash:  "${var.core_user_password}"
      ssh_authorized_keys:
${join("\n", formatlist("        - %s", var.management_ssh_keys))}
        - ${tls_private_key.terraform.public_key_openssh}
    - name: root
      ssh_authorized_keys:
        - ${tls_private_key.terraform.public_key_openssh}
EOF

}

# OS install
resource "null_resource" "controller_nodes_os_install" {
  count = var.controller_nodes_count

  triggers = {
    matchine_id = hcloud_server.controller_nodes[count.index].id
  }

  connection {
    host        = element(hcloud_server.controller_nodes.*.ipv4_address, count.index)
    timeout     = "10m"
    agent       = false
    private_key = tls_private_key.terraform.private_key_pem
  }

  provisioner "file" {
    destination = "/root/ignition.json"
    content     = element(data.ct_config.config.*.rendered, count.index)
  }

  provisioner "remote-exec" {
    inline = [
      "wget -q -O- https://raw.githubusercontent.com/flatcar-linux/init/flatcar-master/bin/flatcar-install | bash -s -- -d /dev/sda -i /root/ignition.json -C edge",
    ]
  }
}

resource "sshcommand_command" "controller_nodes_reboot" {
  count                 = var.controller_nodes_count
  host                  = hcloud_server.controller_nodes[count.index].ipv4_address
  command               = "reboot"
  private_key           = tls_private_key.terraform.private_key_pem
  ignore_execute_errors = true
  depends_on            = [null_resource.controller_nodes_os_install]
}

resource "sshcommand_command" "controller_nodes_wait_for_os" {
  count          = var.controller_nodes_count
  host           = hcloud_server.controller_nodes[count.index].ipv4_address
  command        = "grep ID=flatcar /etc/os-release"
  private_key    = tls_private_key.terraform.private_key_pem
  retry          = true
  retry_interval = "1s"
  depends_on     = [sshcommand_command.controller_nodes_reboot]
}

