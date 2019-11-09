provider "flexkube" {}

module "root_pki" {
  source = "git::https://github.com/flexkube/terraform-root-pki.git"

  organization = "example"
}

module "etcd_pki" {
  source = "git::https://github.com/flexkube/terraform-etcd-pki.git"

  root_ca_cert      = module.root_pki.root_ca_cert
  root_ca_key       = module.root_pki.root_ca_key
  root_ca_algorithm = module.root_pki.root_ca_algorithm

  peer_ips   = ["127.0.0.1"]
  peer_names = ["foo"]

  organization = "example"
}

resource "flexkube_etcd_cluster" "etcd" {
  config = templatefile("./config.yaml.tmpl", {
    peer_ips                  = module.etcd_pki.etcd_peer_ips
    peer_names                = module.etcd_pki.etcd_peer_names
    peer_ca                   = module.etcd_pki.etcd_ca_cert
    peer_certs                = module.etcd_pki.etcd_peer_certs
    peer_keys                 = module.etcd_pki.etcd_peer_keys
  })
}
