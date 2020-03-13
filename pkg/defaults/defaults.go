// Package defaults provides default values used across the library.
package defaults

const (
	// EtcdImage points to a default Docker image, which will be used for running etcd.
	EtcdImage = "quay.io/coreos/etcd:v3.4.4"

	// KubernetesImage is a default container image used for all kubernetes containers.
	KubernetesImage = "k8s.gcr.io/hyperkube:v1.17.4"

	// HAProxyImage is a default container image for APILoadBalancer.
	HAProxyImage = "haproxy:2.1.3-alpine"

	// DockerAPIVersion is a default API version used when talking to Docker runtime.
	DockerAPIVersion = "v1.38"
)
