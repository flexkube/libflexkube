provider "flexkube" {}

module "root_pki" {
  source = "git::https://github.com/invidian/terraform-root-pki.git"

  organization = "example"
}

module "kubernetes_pki" {
  source = "git::https://github.com/invidian/terraform-kubernetes-pki.git"

  root_ca_cert      = module.root_pki.root_ca_cert
  root_ca_key       = module.root_pki.root_ca_key
  root_ca_algorithm = module.root_pki.root_ca_algorithm

  api_server_ips            = ["127.0.0.1"]
  api_server_external_ips   = ["127.0.1.1"]
  api_server_external_names = ["kube-apiserver.example.com"]
  organization = "example"
}

resource "flexkube_controlplane" "controlplane" {
  config = templatefile("./config.yaml.tmpl", {
    kubernetes_ca_certificate = module.kubernetes_pki.kubernetes_ca_cert
    kubernetes_ca_key = module.kubernetes_pki.kubernetes_ca_key
    kubernetes_api_server_certificate = module.kubernetes_pki.kubernetes_api_server_cert
    kubernetes_api_server_key = module.kubernetes_pki.kubernetes_api_server_key
    service_account_public_key = module.kubernetes_pki.service_account_public_key
    service_account_private_key = module.kubernetes_pki.service_account_private_key
    admin_certificate = module.kubernetes_pki.kubernetes_api_server_kubelet_client_cert
    admin_key = module.kubernetes_pki.kubernetes_api_server_kubelet_client_key
    api_server_address = "127.0.0.1"
    etcd_servers = ["http://127.0.0.1:2379"]
    root_ca_certificate = module.root_pki.root_ca_cert
  })
}
