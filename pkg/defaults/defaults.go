package defaults

// use etcd 3.3.17, because 3.4.x is not yet supported
const EtcdImage = "gcr.io/etcd-development/etcd:v3.3.17"
const KubernetesImage = "k8s.gcr.io/hyperkube:v1.16.1"
const HAProxyImage = "haproxy:2.0.7-alpine"
const DockerAPIVersion = "v1.40"
