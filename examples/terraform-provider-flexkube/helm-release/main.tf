provider "flexkube" {}

locals {
  kubeconfig = file("./kubeconfig")
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
}
