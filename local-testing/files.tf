resource "local_file" "kube_apiserver_values" {
  sensitive_content = local.kube_apiserver_values
  filename          = "./values/kube-apiserver.yaml"
}

resource "local_file" "kubernetes_values" {
  sensitive_content = local.kubernetes_values
  filename          = "./values/kubernetes.yaml"
}

resource "local_file" "coredns_values" {
  sensitive_content = local.coredns_values
  filename          = "./values/coredns.yaml"
}

resource "local_file" "calico_values" {
  sensitive_content = local.calico_values
  filename          = "./values/calico.yaml"
}

resource "local_file" "etcd_config" {
  sensitive_content = flexkube_etcd_cluster.etcd.config_yaml
  filename          = "./resources/etcd-cluster/config.yaml"
}

resource "local_file" "etcd_state" {
  sensitive_content = flexkube_etcd_cluster.etcd.state_yaml
  filename          = "./resources/etcd-cluster/state.yaml"
}

resource "local_file" "controlplane_config" {
  sensitive_content = flexkube_controlplane.bootstrap.config_yaml
  filename          = "./resources/controlplane/config.yaml"
}

resource "local_file" "controlplane_state" {
  sensitive_content = flexkube_controlplane.bootstrap.state_yaml
  filename          = "./resources/controlplane/state.yaml"
}

resource "local_file" "apiloadbalancer_config" {
  sensitive_content = flexkube_apiloadbalancer_pool.controllers.config_yaml
  filename          = "./resources/api-loadbalancers/config.yaml"
}

resource "local_file" "apiloadbalancer_state" {
  sensitive_content = flexkube_apiloadbalancer_pool.controllers.state_yaml
  filename          = "./resources/api-loadbalancers/state.yaml"
}

resource "local_file" "kubelet_pool_config" {
  sensitive_content = flexkube_kubelet_pool.controller.config_yaml
  filename          = "./resources/kubelet-pool/config.yaml"
}

resource "local_file" "kubelet_pool_state" {
  sensitive_content = flexkube_kubelet_pool.controller.state_yaml
  filename          = "./resources/kubelet-pool/state.yaml"
}

resource "local_file" "etcd_ca_certificate" {
  content  = flexkube_pki.pki.etcd[0].ca[0].x509_certificate
  filename = "./resources/etcd-cluster/ca.pem"
}

resource "local_file" "etcd_root_user_certificate" {
  content  = flexkube_pki.pki.etcd[0].client_certificates[index(flexkube_pki.pki.etcd[0].client_cns, "root")].x509_certificate
  filename = "./resources/etcd-cluster/client.pem"
}

resource "local_file" "etcd_root_user_private_key" {
  sensitive_content = flexkube_pki.pki.etcd[0].client_certificates[index(flexkube_pki.pki.etcd[0].client_cns, "root")].private_key
  filename          = "./resources/etcd-cluster/client.key"
}

resource "local_file" "etcd_prometheus_user_certificate" {
  content  = flexkube_pki.pki.etcd[0].client_certificates[index(flexkube_pki.pki.etcd[0].client_cns, "prometheus")].x509_certificate
  filename = "./resources/etcd-cluster/prometheus_client.pem"
}

resource "local_file" "etcd_prometheus_user_private_key" {
  sensitive_content = flexkube_pki.pki.etcd[0].client_certificates[index(flexkube_pki.pki.etcd[0].client_cns, "prometheus")].private_key
  filename          = "./resources/etcd-cluster/prometheus_client.key"
}

resource "local_file" "etcd_environment" {
  filename = "./resources/etcd-cluster/environment.sh"
  content  = <<EOF
#!/bin/bash
export ETCDCTL_API=3
export ETCDCTL_CACERT=${abspath(local_file.etcd_ca_certificate.filename)}
export ETCDCTL_CERT=${abspath(local_file.etcd_root_user_certificate.filename)}
export ETCDCTL_KEY=${abspath(local_file.etcd_root_user_private_key.filename)}
export ETCDCTL_ENDPOINTS=${join(",", local.etcd_servers)}
EOF

  depends_on = [
    flexkube_etcd_cluster.etcd,
  ]
}

resource "local_file" "etcd_prometheus_environment" {
  filename = "./resources/etcd-cluster/prometheus-environment.sh"
  content  = <<EOF
#!/bin/bash
export ETCDCTL_API=3
export ETCDCTL_CACERT=${abspath(local_file.etcd_ca_certificate.filename)}
export ETCDCTL_CERT=${abspath(local_file.etcd_prometheus_user_certificate.filename)}
export ETCDCTL_KEY=${abspath(local_file.etcd_prometheus_user_private_key.filename)}
export ETCDCTL_ENDPOINTS=${join(",", local.etcd_servers)}
EOF

  depends_on = [
    flexkube_etcd_cluster.etcd,
  ]
}

resource "local_file" "etcd_enable_rbac" {
  filename = "./resources/etcd-cluster/enable-rbac.sh"
  content  = <<EOF
#!/bin/bash
etcdctl user add --no-password=true root
etcdctl role add root
etcdctl user grant-role root root
etcdctl auth enable
etcdctl user add --no-password=true kube-apiserver
etcdctl role add kube-apiserver
etcdctl role grant-permission kube-apiserver readwrite --prefix=true /
etcdctl user grant-role kube-apiserver kube-apiserver
# Until https://github.com/etcd-io/etcd/issues/8458 is resolved.
etcdctl user grant-role kube-apiserver root
etcdctl user add --no-password=true prometheus
EOF
}
