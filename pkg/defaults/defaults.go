// Package defaults provides default values used across the library.
package defaults

const (
	// EtcdImage points to a default Docker image, which will be used for running etcd.
	EtcdImage = "quay.io/coreos/etcd:v3.5.4"

	// KubeAPIServerImage points to a default Docker image, which will be used for
	// running kube-apiserver.
	KubeAPIServerImage = "k8s.gcr.io/kube-apiserver:v1.23.5"

	// KubeControllerManagerImage points to a default Docker image, which will be used for
	// running kube-apiserver.
	KubeControllerManagerImage = "k8s.gcr.io/kube-controller-manager:v1.23.5"

	// KubeSchedulerImage points to a default Docker image, which will be used for
	// running kube-apiserver.
	KubeSchedulerImage = "k8s.gcr.io/kube-scheduler:v1.23.5"

	// KubeletImage points to a default Docker image, which will be used for
	// running kube-apiserver.
	KubeletImage = "quay.io/flexkube/kubelet:v1.23.5"

	// HAProxyImage is a default container image for APILoadBalancer.
	HAProxyImage = "haproxy:2.5.5-alpine"

	// DockerAPIVersion is a default API version used when talking to Docker runtime.
	DockerAPIVersion = "v1.38"

	// VolumePluginDir is a default flex volume plugin directory configured for kubelet
	// and kube-controller-manager.
	VolumePluginDir = "/usr/libexec/kubernetes/kubelet-plugins/volume/exec"
)
