package defaults

// EtcdImage points to a default Docker image, which will be used for running etcd.
// We use etcd 3.3.17, because 3.4.x is not yet supported
const EtcdImage = "gcr.io/etcd-development/etcd:v3.3.17"

// KubernetesImage is a default container image used for all kubernetes containers
const KubernetesImage = "k8s.gcr.io/hyperkube:v1.16.3"

// HAProxyImage is a default container image for APILoadBalancer
const HAProxyImage = "haproxy:2.0.7-alpine"

// DockerAPIVersion is a default API version used when talking to Docker runtime
const DockerAPIVersion = "v1.40"
