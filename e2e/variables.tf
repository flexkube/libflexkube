variable "ssh_private_key_path" {
  default = "/root/.ssh/id_rsa"
}

variable "controllers_count" {
  default = 1
}

variable "workers_count" {
  default = 0
}

variable "nodes_cidr" {
  default = "192.168.50.0/24"
}

variable "pod_cidr" {
  default = "10.1.0.0/16"
}

variable "network_plugin" {
  default = "calico"
}

variable "node_ssh_port" {
  default = 22
}

variable "kube_apiserver_helm_chart_source" {
  default = "flexkube/kube-apiserver"
}

variable "kubernetes_helm_chart_source" {
  default = "flexkube/kubernetes"
}

variable "kube_proxy_helm_chart_source" {
  default = "flexkube/kube-proxy"
}

variable "tls_bootstrapping_helm_chart_source" {
  default = "flexkube/tls-bootstrapping"
}

variable "kubelet_rubber_stamp_helm_chart_source" {
  default = "flexkube/kubelet-rubber-stamp"
}

variable "calico_helm_chart_source" {
  default = "flexkube/calico"
}

variable "kube_apiserver_helm_chart_version" {
  default = "0.1.14"
}

variable "kubernetes_helm_chart_version" {
  default = "0.2.3"
}

variable "kube_proxy_helm_chart_version" {
  default = "0.2.0"
}

variable "tls_bootstrapping_helm_chart_version" {
  default = "0.1.1"
}

variable "coredns_chart_version" {
  default = "1.13.3"
}

variable "metrics_server_chart_version" {
  default = "2.11.1"
}

variable "kubelet_rubber_stamp_helm_chart_version" {
  default = "0.1.4"
}

variable "calico_helm_chart_version" {
  default = "0.2.3"
}

variable "flatcar_channel" {
  default = "edge"
}
