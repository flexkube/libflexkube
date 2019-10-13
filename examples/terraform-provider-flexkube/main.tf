provider "local" {
  version = "~> 1.3"
}
provider "flexkube" {}

variable "hcloud_token" {}
variable "domain" {
  default = "example.com"
}

variable "management_cidrs" {
  default = "0.0.0.0/0"
}

variable "management_ssh_keys" {
  default = []
}

module "hetzner_machines" {
  source = "./hetzner_machines"

  domain       = var.domain
  hcloud_token = var.hcloud_token

  controller_nodes_count = 2
  management_cidrs    = var.management_cidrs
  management_ssh_keys = var.management_ssh_keys
}

module "root_pki" {
  source = "./root_pki"
}

module "etcd_pki" {
  source = "./etcd_pki"

  root_ca_cert      = module.root_pki.root_ca_cert
  root_ca_key       = module.root_pki.root_ca_key
  root_ca_algorithm = module.root_pki.root_ca_algorithm

  peer_ips   = module.hetzner_machines.node_wireguard_ips
  peer_names = module.hetzner_machines.node_names
}

resource "flexkube_etcd_cluster" "foo" {
  config = templatefile("./config.yaml.tmpl", {
    peer_ssh_addresses = module.hetzner_machines.node_public_ips
    peer_ips           = module.etcd_pki.etcd_peer_ips
    peer_names         = module.etcd_pki.etcd_peer_names
    peer_ca            = module.etcd_pki.etcd_ca_cert
    peer_certs         = module.etcd_pki.etcd_peer_certs
    peer_keys          = module.etcd_pki.etcd_peer_keys
    ssh_private_key    = module.hetzner_machines.provisioning_private_key
  })

  depends_on = [
    "module.hetzner_machines"
  ]
}

resource "local_file" "config" {
  sensitive_content = flexkube_etcd_cluster.foo.config
  filename          = "./config.yaml"
}

resource "local_file" "state" {
  sensitive_content = flexkube_etcd_cluster.foo.state
  filename          = "./state.yaml"
}
