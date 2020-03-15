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
  sensitive_content = flexkube_apiloadbalancer_pool.nodes.config_yaml
  filename          = "./resources/api-loadbalancers/config.yaml"
}

resource "local_file" "apiloadbalancer_state" {
  sensitive_content = flexkube_apiloadbalancer_pool.nodes.state_yaml
  filename          = "./resources/api-loadbalancers/state.yaml"
}

resource "local_file" "kubelet_pool_config" {
  sensitive_content = local.kubelet_pool_config
  filename          = "./resources/kubelet-pool/config.yaml"
}

resource "local_file" "kubelet_pool_state" {
  sensitive_content = flexkube_kubelet_pool.controller.state
  filename          = "./resources/kubelet-pool/state.yaml"
}
