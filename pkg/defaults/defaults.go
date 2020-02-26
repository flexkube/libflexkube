package defaults

const (
	// EtcdImage points to a default Docker image, which will be used for running etcd.
	EtcdImage = "gcr.io/etcd-development/etcd:v3.4.4"

	// KubernetesImage is a default container image used for all kubernetes containers.
	KubernetesImage = "k8s.gcr.io/hyperkube:v1.17.3"

	// HAProxyImage is a default container image for APILoadBalancer.
	HAProxyImage = "haproxy:2.1.2-alpine"

	// DockerAPIVersion is a default API version used when talking to Docker runtime.
	DockerAPIVersion = "v1.38"
)
