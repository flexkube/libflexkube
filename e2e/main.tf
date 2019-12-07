terraform {
  required_version = "~> 0.12.0"
}

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

  peer_ips   = ["10.0.2.15"]
  peer_names = ["foo"]

  organization = "example"
}

module "kubernetes_pki" {
  source = "git::https://github.com/flexkube/terraform-kubernetes-pki.git"

  root_ca_cert      = module.root_pki.root_ca_cert
  root_ca_key       = module.root_pki.root_ca_key
  root_ca_algorithm = module.root_pki.root_ca_algorithm

  api_server_ips            = ["10.0.2.15"]
  api_server_external_ips   = ["127.0.1.1"]
  api_server_external_names = ["kube-apiserver.example.com"]
  organization = "example"
}

locals {
  kubeconfig = templatefile("./templates/kubeconfig.tmpl", {
    name        = "e2e"
    server      = "https://10.0.2.15:6443"
    ca_cert     = base64encode(module.kubernetes_pki.kubernetes_ca_cert)
    client_cert = base64encode(module.kubernetes_pki.kubernetes_api_server_kubelet_client_cert)
    client_key  = base64encode(module.kubernetes_pki.kubernetes_api_server_kubelet_client_key)
  })

  kubeconfig_host = templatefile("./templates/kubeconfig.tmpl", {
    name        = "e2e"
    server      = "https://127.0.0.1:6443"
    ca_cert     = base64encode(module.kubernetes_pki.kubernetes_ca_cert)
    client_cert = base64encode(module.kubernetes_pki.kubernetes_api_server_kubelet_client_cert)
    client_key  = base64encode(module.kubernetes_pki.kubernetes_api_server_kubelet_client_key)
  })

  kubernetes_values = templatefile("./templates/values.yaml.tmpl", {
    kubernetes_api_server_key = module.kubernetes_pki.kubernetes_api_server_key
    kubernetes_api_server_certificate = module.kubernetes_pki.kubernetes_api_server_cert
    service_account_public_key = module.kubernetes_pki.service_account_public_key
    kubernetes_ca_certificate = module.kubernetes_pki.kubernetes_ca_cert
    service_account_private_key = module.kubernetes_pki.service_account_private_key
    kubernetes_ca_key = module.kubernetes_pki.kubernetes_ca_key
    etcd_servers = formatlist("http://%s:2379", module.etcd_pki.etcd_peer_ips)
    root_ca_certificate = module.root_pki.root_ca_cert
    replicas = 1
    max_unavailable = 0
  })

  bootstrap_kubeconfig = templatefile("./templates/bootstrap-kubeconfig.tmpl", {
    address = "10.0.2.15"
  })
}

resource "local_file" "kubeconfig" {
  sensitive_content = local.kubeconfig_host
  filename          = "./kubeconfig"
}

resource "flexkube_etcd_cluster" "etcd" {
  config = templatefile("./templates/etcd_config.yaml.tmpl", {
    peer_ssh_addresses        = ["10.0.2.15"]
    peer_ips                  = module.etcd_pki.etcd_peer_ips
    peer_names                = module.etcd_pki.etcd_peer_names
    peer_ca                   = module.etcd_pki.etcd_ca_cert
    peer_certs                = module.etcd_pki.etcd_peer_certs
    peer_keys                 = module.etcd_pki.etcd_peer_keys
    ssh_private_key           = file("/root/.ssh/id_rsa")
  })
}

resource "flexkube_apiloadbalancer_pool" "controllers" {
  count = 0
  config = templatefile("./templates/apiloadbalancer_pool_config.yaml.tmpl", {
    metrics_bind_addresses = ["10.0.2.15"]
    ssh_private_key = file("/root/.ssh/id_rsa")
    ssh_addresses = ["10.0.2.15"]
    servers = ["10.0.2.15"]
  })

  depends_on = [
    flexkube_etcd_cluster.etcd,
  ]
}

resource "flexkube_controlplane" "bootstrap" {
  config = templatefile("./templates/controlplane_config.yaml.tmpl", {
    kubernetes_ca_certificate = module.kubernetes_pki.kubernetes_ca_cert
    kubernetes_ca_key = module.kubernetes_pki.kubernetes_ca_key
    kubernetes_api_server_certificate = module.kubernetes_pki.kubernetes_api_server_cert
    kubernetes_api_server_key = module.kubernetes_pki.kubernetes_api_server_key
    service_account_public_key = module.kubernetes_pki.service_account_public_key
    service_account_private_key = module.kubernetes_pki.service_account_private_key
    admin_certificate = module.kubernetes_pki.kubernetes_api_server_kubelet_client_cert
    admin_key = module.kubernetes_pki.kubernetes_api_server_kubelet_client_key
    api_server_address = "10.0.2.15"
    etcd_servers = formatlist("http://%s:2379", module.etcd_pki.etcd_peer_ips)
    ssh_address = "10.0.2.15"
    ssh_private_key = file("/root/.ssh/id_rsa")
    root_ca_certificate = module.root_pki.root_ca_cert
    replicas = 1
  })

  depends_on = [
    flexkube_apiloadbalancer_pool.controllers,
    flexkube_etcd_cluster.etcd,
  ]
}

resource "flexkube_helm_release" "coredns" {
  kubeconfig = local.kubeconfig
  namespace  = "kube-system"
  chart      = "stable/coredns"
  name       = "coredns"
  values     = <<EOF
service:
  clusterIP: 11.0.0.10
EOF

  depends_on = [
    flexkube_controlplane.bootstrap
  ]
}

resource "flexkube_helm_release" "kubernetes" {
  kubeconfig = local.kubeconfig
  namespace  = "kube-system"
  chart      = "/usr/src/libflexkube/charts/kubernetes"
  name       = "kubernetes"
  values     = local.kubernetes_values

  depends_on = [
    flexkube_controlplane.bootstrap
  ]
}

resource "flexkube_helm_release" "kubelet-rubber-stamp" {
  kubeconfig = local.kubeconfig
  namespace  = "kube-system"
  chart      = "/usr/src/libflexkube/charts/kubelet-rubber-stamp"
  name       = "kubelet-rubber-stamp"

  depends_on = [
    flexkube_controlplane.bootstrap
  ]
}

resource "flexkube_kubelet_pool" "controller" {
  config = templatefile("./templates/kubelet_config.yaml.tmpl", {
    kubelet_addresses         = ["10.0.2.15"]
    bootstrap_kubeconfigs     = [local.bootstrap_kubeconfig]
    ssh_private_key           = file("/root/.ssh/id_rsa")
    ssh_addresses             = ["10.0.2.15"]
    kubelet_pod_cidrs         = ["10.1.0.0/24"]
    kubernetes_ca_certificate = module.kubernetes_pki.kubernetes_ca_cert
  })

  depends_on = [
    flexkube_apiloadbalancer_pool.controllers,
    flexkube_helm_release.kubernetes,
  ]
}
