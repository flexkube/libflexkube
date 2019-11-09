provider "flexkube" {}

module "root_pki" {
  source = "git::https://github.com/flexkube/terraform-root-pki.git"

  organization = "example"
}

module "kubernetes_pki" {
  source = "git::https://github.com/flexkube/terraform-kubernetes-pki.git"

  root_ca_cert      = module.root_pki.root_ca_cert
  root_ca_key       = module.root_pki.root_ca_key
  root_ca_algorithm = module.root_pki.root_ca_algorithm

  api_server_ips            = ["127.0.0.1"]
  api_server_external_ips   = ["127.0.1.1"]
  api_server_external_names = ["kube-apiserver.example.com"]
  organization = "example"
}

resource "flexkube_kubelet_pool" "controller" {
  config = templatefile("./config.yaml.tmpl", {
    kubelet_addresses         = ["127.0.0.1"]
    kubelet_pod_cidrs         = ["10.42.0.0/24"]
    kubernetes_ca_certificate = module.kubernetes_pki.kubernetes_ca_cert
  })
}
