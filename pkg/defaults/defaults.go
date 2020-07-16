// Package defaults provides default values used across the library.
package defaults

const (
	// EtcdImage points to a default Docker image, which will be used for running etcd.
	EtcdImage = "quay.io/coreos/etcd:v3.4.9"

	// KubernetesImage is a default container image used for all kubernetes containers.
	KubernetesImage = "k8s.gcr.io/hyperkube:v1.18.6"

	// HAProxyImage is a default container image for APILoadBalancer.
	HAProxyImage = "haproxy:2.1.7-alpine"

	// DockerAPIVersion is a default API version used when talking to Docker runtime.
	DockerAPIVersion = "v1.38"

	// VolumePluginDir is a default flex volume plugin directory configured for kubelet
	// and kube-controller-manager.
	VolumePluginDir = "/usr/libexec/kubernetes/kubelet-plugins/volume/exec"
)
