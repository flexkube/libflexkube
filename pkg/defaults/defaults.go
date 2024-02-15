// Package defaults provides default values used across the library.
package defaults

const (
	// EtcdImage points to a default Docker image, which will be used for running etcd.
	EtcdImage = "quay.io/coreos/etcd:v3.5.12"

	// KubeAPIServerImage points to a default Docker image, which will be used for
	// running kube-apiserver.
	KubeAPIServerImage = "registry.k8s.io/kube-apiserver:v1.29.2"

	// KubeControllerManagerImage points to a default Docker image, which will be used for
	// running kube-apiserver.
	KubeControllerManagerImage = "registry.k8s.io/kube-controller-manager:v1.29.2"

	// KubeSchedulerImage points to a default Docker image, which will be used for
	// running kube-apiserver.
	KubeSchedulerImage = "registry.k8s.io/kube-scheduler:v1.29.2"

	// KubeletImage points to a default Docker image, which will be used for
	// running kube-apiserver.
	KubeletImage = "quay.io/flexkube/kubelet:v1.29.2"

	// HAProxyImage is a default container image for APILoadBalancer.
	HAProxyImage = "haproxy:2.9.4-alpine"

	// DockerAPIVersion is a default API version used when talking to Docker runtime.
	DockerAPIVersion = "v1.38"

	// VolumePluginDir is a default flex volume plugin directory configured for kubelet
	// and kube-controller-manager.
	VolumePluginDir = "/usr/libexec/kubernetes/kubelet-plugins/volume/exec"
)
