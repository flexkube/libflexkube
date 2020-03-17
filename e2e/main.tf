provider "local" {
  version = "= 1.4.0"
}

module "root_pki" {
  source = "git::https://github.com/flexkube/terraform-root-pki.git"

  organization = "example"
}

module "etcd_pki" {
  source = "git::https://github.com/flexkube/terraform-etcd-pki.git"

  root_ca_cert      = module.root_pki.root_ca_cert
  root_ca_key       = module.root_pki.root_ca_key
  root_ca_algorithm = module.root_pki.root_ca_algorithm

  peer_ips   = local.controller_ips
  peer_names = local.controller_names

  server_ips   = local.controller_ips
  server_names = local.controller_names

  client_cns = ["kube-apiserver-etcd-client"]

  organization = "example"
}

module "kubernetes_pki" {
  source = "git::https://github.com/flexkube/terraform-kubernetes-pki.git"

  root_ca_cert      = module.root_pki.root_ca_cert
  root_ca_key       = module.root_pki.root_ca_key
  root_ca_algorithm = module.root_pki.root_ca_algorithm

  api_server_ips            = local.controller_ips
  api_server_external_ips   = ["127.0.1.1"]
  api_server_external_names = ["kube-apiserver.example.com"]
  organization              = "example"
}

locals {
  cgroup_driver = var.flatcar_channel == "edge" ? "systemd" : "cgroupfs"

  bootstrap_api_bind = "127.0.0.1"
  api_port           = 8443

  node_load_balancer_address = "127.0.0.1:7443"

  kube_apiserver_values = templatefile("./templates/kube-apiserver-values.yaml.tmpl", {
    server_key                     = module.kubernetes_pki.kubernetes_api_server_key
    server_certificate             = module.kubernetes_pki.kubernetes_api_server_cert
    service_account_public_key     = module.kubernetes_pki.service_account_public_key
    ca_certificate                 = module.kubernetes_pki.kubernetes_ca_cert
    front_proxy_client_key         = module.kubernetes_pki.kubernetes_api_server_front_proxy_client_key
    front_proxy_client_certificate = module.kubernetes_pki.kubernetes_api_server_front_proxy_client_cert
    front_proxy_ca_certificate     = module.kubernetes_pki.kubernetes_front_proxy_ca_cert
    kubelet_client_certificate     = module.kubernetes_pki.kubernetes_api_server_kubelet_client_cert
    kubelet_client_key             = module.kubernetes_pki.kubernetes_api_server_kubelet_client_key
    etcd_ca_certificate            = module.etcd_pki.etcd_ca_cert
    etcd_client_certificate        = module.etcd_pki.client_certs[0]
    etcd_client_key                = module.etcd_pki.client_keys[0]
    etcd_servers                   = formatlist("https://%s:2379", module.etcd_pki.etcd_peer_ips)
    replicas                       = var.controllers_count
  })

  kubernetes_values = templatefile("./templates/values.yaml.tmpl", {
    service_account_private_key = module.kubernetes_pki.service_account_private_key
    kubernetes_ca_key           = module.kubernetes_pki.kubernetes_ca_key
    root_ca_certificate         = module.root_pki.root_ca_cert
    kubernetes_ca_certificate   = module.kubernetes_pki.kubernetes_ca_cert
    api_servers                 = formatlist("%s:%d", local.controller_ips, local.api_port)
    replicas                    = var.controllers_count
    podsCIDR                    = var.pod_cidr
  })

  coredns_values = <<EOF
rbac:
  pspEnable: true
service:
  clusterIP: 11.0.0.10
nodeSelector:
  node-role.kubernetes.io/master: ""
tolerations:
  - key: node-role.kubernetes.io/master
    operator: Exists
    effect: NoSchedule
EOF

  calico_values = <<EOF
podCIDR: ${var.pod_cidr}
flexVolumePluginDir: /var/lib/kubelet/volumeplugins
EOF

  metrics_server_values = <<EOF
rbac:
  pspEnabled: true
args:
- --kubelet-preferred-address-types=InternalIP
podDisruptionBudget:
  enabled: true
  minAvailable: 1
tolerations:
- key: node-role.kubernetes.io/master
  operator: Exists
  effect: NoSchedule
resources:
  requests:
    memory: 20Mi
EOF

  kubeconfig_admin = templatefile("./templates/kubeconfig.tmpl", {
    name        = "admin"
    server      = "https://${local.first_controller_ip}:${local.api_port}"
    ca_cert     = base64encode(module.kubernetes_pki.kubernetes_ca_cert)
    client_cert = base64encode(module.kubernetes_pki.kubernetes_admin_cert)
    client_key  = base64encode(module.kubernetes_pki.kubernetes_admin_key)
  })

  network_plugin = var.network_plugin == "kubenet" ? "kubenet" : "cni"

  bootstrap_kubeconfig = <<EOF
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority: /etc/kubernetes/pki/ca.crt
    server: https://${local.node_load_balancer_address}
  name: bootstrap
contexts:
- context:
    cluster: bootstrap
    user: kubelet-bootstrap
  name: bootstrap
current-context: bootstrap
preferences: {}
users:
- name: kubelet-bootstrap
  user:
    token: 07401b.f395accd246ae52d
EOF

  deploy_workers = var.workers_count > 0 ? 1 : 0

  ssh_private_key = file(var.ssh_private_key_path)
}

resource "local_file" "kubeconfig" {
  sensitive_content = local.kubeconfig_admin
  filename          = "./kubeconfig"
}

resource "flexkube_etcd_cluster" "etcd" {
  ssh {
    user        = "core"
    port        = var.node_ssh_port
    private_key = local.ssh_private_key
  }

  ca_certificate = module.etcd_pki.etcd_ca_cert

  dynamic "member" {
    for_each = module.etcd_pki.etcd_peer_ips

    content {
      name               = module.etcd_pki.etcd_peer_names[member.key]
      peer_certificate   = module.etcd_pki.etcd_peer_certs[member.key]
      peer_key           = module.etcd_pki.etcd_peer_keys[member.key]
      server_certificate = module.etcd_pki.etcd_server_certs[member.key]
      server_key         = module.etcd_pki.etcd_server_keys[member.key]
      peer_address       = module.etcd_pki.etcd_peer_ips[member.key]
      server_address     = local.controller_ips[member.key]

      host {
        ssh {
          address = local.controller_ips[member.key]
        }
      }
    }
  }
}

resource "flexkube_apiloadbalancer_pool" "nodes" {
  name             = "api-loadbalancer-nodes"
  host_config_path = "/etc/haproxy/nodes.cfg"
  bind_address     = local.node_load_balancer_address
  servers          = formatlist("%s:%d", local.controller_ips, local.api_port)

  ssh {
    private_key = local.ssh_private_key
    port        = var.node_ssh_port
    user        = "core"
  }

  dynamic "api_load_balancer" {
    for_each = concat(local.controller_ips, local.worker_ips)

    content {
      host {
        ssh {
          address = api_load_balancer.value
        }
      }
    }
  }
}

resource "flexkube_apiloadbalancer_pool" "bootstrap" {
  name             = "api-loadbalancer-bootstrap"
  host_config_path = "/etc/haproxy/bootstrap.cfg"
  bind_address     = "${local.first_controller_ip}:${local.api_port}"
  servers          = ["${local.bootstrap_api_bind}:${local.api_port}"]

  ssh {
    private_key = local.ssh_private_key
    port        = var.node_ssh_port
    user        = "core"
  }

  api_load_balancer {
    host {
      ssh {
        address = local.first_controller_ip
      }
    }
  }
}

resource "flexkube_controlplane" "bootstrap" {
  common {
    kubernetes_ca_certificate  = module.kubernetes_pki.kubernetes_ca_cert
    front_proxy_ca_certificate = module.kubernetes_pki.kubernetes_front_proxy_ca_cert
  }

  kube_apiserver {
    api_server_certificate     = module.kubernetes_pki.kubernetes_api_server_cert
    api_server_key             = module.kubernetes_pki.kubernetes_api_server_key
    front_proxy_certificate    = module.kubernetes_pki.kubernetes_api_server_front_proxy_client_cert
    front_proxy_key            = module.kubernetes_pki.kubernetes_api_server_front_proxy_client_key
    kubelet_client_certificate = module.kubernetes_pki.kubernetes_api_server_kubelet_client_cert
    kubelet_client_key         = module.kubernetes_pki.kubernetes_api_server_kubelet_client_key
    service_account_public_key = module.kubernetes_pki.service_account_public_key
    etcd_ca_certificate        = module.etcd_pki.etcd_ca_cert
    etcd_client_certificate    = module.etcd_pki.client_certs[0]
    etcd_client_key            = module.etcd_pki.client_keys[0]
    service_cidr               = "11.0.0.0/24"
    etcd_servers               = formatlist("https://%s:2379", module.etcd_pki.etcd_peer_ips)
    bind_address               = local.bootstrap_api_bind
    advertise_address          = local.first_controller_ip
    secure_port                = local.api_port
  }

  kube_controller_manager {
    flex_volume_plugin_dir      = "/var/lib/kubelet/volumeplugins"
    kubernetes_ca_key           = module.kubernetes_pki.kubernetes_ca_key
    service_account_private_key = module.kubernetes_pki.service_account_private_key
    root_ca_certificate         = module.root_pki.root_ca_cert

    kubeconfig {
      client_certificate = module.kubernetes_pki.kube_controller_manager_cert
      client_key         = module.kubernetes_pki.kube_controller_manager_key
    }
  }

  kube_scheduler {
    kubeconfig {
      client_certificate = module.kubernetes_pki.kube_scheduler_cert
      client_key         = module.kubernetes_pki.kube_scheduler_key
    }
  }

  api_server_address = local.first_controller_ip
  api_server_port    = local.api_port

  ssh {
    user        = "core"
    address     = local.first_controller_ip
    port        = var.node_ssh_port
    private_key = local.ssh_private_key
  }

  depends_on = [
    flexkube_etcd_cluster.etcd,
  ]
}

resource "flexkube_helm_release" "kube-apiserver" {
  kubeconfig = local.kubeconfig_admin
  namespace  = "kube-system"
  chart      = var.kube_apiserver_helm_chart_source
  name       = "kube-apiserver"
  values     = local.kube_apiserver_values

  depends_on = [
    flexkube_controlplane.bootstrap,
    flexkube_apiloadbalancer_pool.nodes,
    flexkube_apiloadbalancer_pool.bootstrap,
  ]
}

resource "flexkube_helm_release" "kubernetes" {
  kubeconfig = local.kubeconfig_admin
  namespace  = "kube-system"
  chart      = var.kubernetes_helm_chart_source
  name       = "kubernetes"
  values     = local.kubernetes_values

  depends_on = [
    flexkube_helm_release.kube-apiserver,
  ]
}

resource "flexkube_helm_release" "coredns" {
  kubeconfig = local.kubeconfig_admin
  namespace  = "kube-system"
  chart      = "stable/coredns"
  name       = "coredns"
  values     = local.coredns_values

  depends_on = [
    flexkube_helm_release.kubernetes,
  ]
}

resource "flexkube_helm_release" "metrics-server" {
  kubeconfig = local.kubeconfig_admin
  namespace  = "kube-system"
  chart      = "stable/metrics-server"
  name       = "metrics-server"
  values     = local.metrics_server_values

  depends_on = [
    flexkube_helm_release.kubernetes,
  ]
}

resource "flexkube_helm_release" "kubelet-rubber-stamp" {
  kubeconfig = local.kubeconfig_admin
  namespace  = "kube-system"
  chart      = var.kubelet_rubber_stamp_helm_chart_source
  name       = "kubelet-rubber-stamp"

  depends_on = [
    flexkube_helm_release.kubernetes,
  ]
}

resource "flexkube_helm_release" "calico" {
  count = var.network_plugin == "calico" ? 1 : 0

  kubeconfig = local.kubeconfig_admin
  namespace  = "kube-system"
  chart      = var.calico_helm_chart_source
  name       = "calico"
  values     = local.calico_values

  depends_on = [
    flexkube_helm_release.kubernetes,
  ]
}

resource "flexkube_kubelet_pool" "controller" {
  bootstrap_kubeconfig      = local.bootstrap_kubeconfig
  cgroup_driver             = local.cgroup_driver
  network_plugin            = local.network_plugin
  kubernetes_ca_certificate = module.kubernetes_pki.kubernetes_ca_cert
  hairpin_mode              = local.network_plugin == "kubenet" ? "promiscuous-bridge" : "hairpin-veth"
  volume_plugin_dir         = "/var/lib/kubelet/volumeplugins"
  cluster_dns_ips = [
    "11.0.0.10"
  ]

  system_reserved = {
    "cpu"    = "100m"
    "memory" = "500Mi"
  }

  kube_reserved = {
    // 100MB for kubelet and 200MB for etcd.
    "memory" = "300Mi"
    "cpu"    = "100m"
  }

  privileged_labels = {
    "node-role.kubernetes.io/master" = ""
  }

  privileged_labels_kubeconfig = local.kubeconfig_admin

  taints = {
    "node-role.kubernetes.io/master" = "NoSchedule"
  }

  ssh {
    user        = "core"
    port        = var.node_ssh_port
    private_key = local.ssh_private_key
  }

  dynamic "kubelet" {
    for_each = local.controller_ips

    content {
      name     = local.controller_names[kubelet.key]
      pod_cidr = local.network_plugin == "kubenet" ? local.controller_cidrs[kubelet.key] : ""

      address = local.controller_ips[kubelet.key]

      host {
        ssh {
          address = kubelet.value
        }
      }
    }
  }

  depends_on = [
    flexkube_apiloadbalancer_pool.nodes,
    flexkube_helm_release.kubernetes,
    flexkube_apiloadbalancer_pool.bootstrap,
  ]
}

resource "flexkube_kubelet_pool" "workers" {
  count = local.deploy_workers

  bootstrap_kubeconfig      = local.bootstrap_kubeconfig
  cgroup_driver             = local.cgroup_driver
  network_plugin            = local.network_plugin
  kubernetes_ca_certificate = module.kubernetes_pki.kubernetes_ca_cert
  hairpin_mode              = local.network_plugin == "kubenet" ? "promiscuous-bridge" : "hairpin-veth"
  volume_plugin_dir         = "/var/lib/kubelet/volumeplugins"
  cluster_dns_ips = [
    "11.0.0.10"
  ]

  system_reserved = {
    "cpu"    = "100m"
    "memory" = "500Mi"
  }

  kube_reserved = {
    // 100MB for kubelet and 200MB for etcd.
    "memory" = "300Mi"
    "cpu"    = "100m"
  }

  ssh {
    user        = "core"
    port        = var.node_ssh_port
    private_key = local.ssh_private_key
  }

  dynamic "kubelet" {
    for_each = local.worker_ips

    content {
      name     = local.worker_names[kubelet.key]
      pod_cidr = local.network_plugin == "kubenet" ? local.worker_cidrs[kubelet.key] : ""

      address = local.worker_ips[kubelet.key]

      host {
        ssh {
          address = kubelet.value
        }
      }
    }
  }

  depends_on = [
    flexkube_apiloadbalancer_pool.nodes,
    flexkube_helm_release.kubernetes,
    flexkube_apiloadbalancer_pool.bootstrap,
  ]
}
