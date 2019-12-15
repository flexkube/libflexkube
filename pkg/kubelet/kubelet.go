package kubelet

import (
	"fmt"
	"strings"

	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/defaults"
	"github.com/flexkube/libflexkube/pkg/host"
)

// Kubelet represents single kubelet instance
type Kubelet struct {
	Address             string     `json:"address,omitempty" yaml:"address,omitempty"`
	Image               string     `json:"image,omitempty" yaml:"image,omitempty"`
	Host                *host.Host `json:"host,omitempty" yaml:"host,omitempty"`
	BootstrapKubeconfig string     `json:"bootstrapKubeconfig,omitempty" yaml:"bootstrapKubeconfig,omitempty"`
	// TODO we require CA certificate, so it can be referred in bootstrap-kubeconfig. Maybe we should be responsible for creating
	// bootstrap-kubeconfig too then?
	KubernetesCACertificate string   `json:"kubernetesCACertificate,omitempty" yaml:"kubernetesCACertificate,omitempty"`
	ClusterDNSIPs           []string `json:"clusterDNSIPs,omitempty" yaml:"clusterDNSIPs,omitempty"`

	// Depending on the network plugin, this should be optional, but for now it's required.
	PodCIDR string `json:"podCIDR,omitempty" yaml:"podCIDR,omitempty"`
}

// kubelet is a validated, executable version of Kubelet
type kubelet struct {
	address                 string
	image                   string
	host                    *host.Host
	bootstrapKubeconfig     string
	kubernetesCACertificate string
	clusterDNSIPs           []string
	podCIDR                 string
}

// New validates Kubelet configuration and returns it's usable version
func (k *Kubelet) New() (*kubelet, error) {
	// TODO when creating kubelet, also pull pause image using configured Container Runtime to speed up later start of pods?
	if err := k.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate kubelet configuration: %w", err)
	}

	nk := &kubelet{
		image:                   k.Image,
		address:                 k.Address,
		host:                    k.Host,
		bootstrapKubeconfig:     k.BootstrapKubeconfig,
		kubernetesCACertificate: k.KubernetesCACertificate,
		clusterDNSIPs:           k.ClusterDNSIPs,
		podCIDR:                 k.PodCIDR,
	}

	if nk.image == "" {
		nk.image = defaults.KubernetesImage
	}

	return nk, nil
}

// Validate validates kubelet configuration
//
// TODO better validation should be done here
func (k *Kubelet) Validate() error {
	if k.BootstrapKubeconfig == "" {
		return fmt.Errorf("bootstrapKubeconfig can't be empty")
	}

	return nil
}

// ToHostConfiguredContainer takes configured kubelet and converts it to generic HostConfiguredContainer
func (k *kubelet) ToHostConfiguredContainer() *container.HostConfiguredContainer {
	configFiles := make(map[string]string)

	// TODO we should use proper templating engine or marshalling for those values.
	clusterDNS := ""
	for _, e := range k.clusterDNSIPs {
		clusterDNS = fmt.Sprintf("%s- %s\n", clusterDNS, e)
	}

	// kubelet.yaml file is a recommended way to configure the kubelet
	//
	// TODO maybe we store that as a struct, and we marshal here to YAML?
	configFiles["/etc/kubernetes/kubelet/kubelet.yaml"] = fmt.Sprintf(`apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
# Enables TLS certificate rotation, which is good from security point of view.
rotateCertificates: true
# Request HTTPS server certs from API as well, so kubelet does not generate self-signed certificates
serverTLSBootstrap: true
# To address: "--cgroups-per-qos enabled, but --cgroup-root was not specified.  defaulting to /"
# This disables QoS based cgroup hierarchy, which is important from resource management perspective.
cgroupsPerQOS: false
# When cgroupsPerQOS is false, enforceNodeAllocatable needs to be set explicitly to empty.
enforceNodeAllocatable: []
# If Docker is configured to use systemd as a cgroup driver and Docker is used as container
# runtime, this needs to be set to match Docker.
# TODO pull that information dynamically based on what container runtime is configured.
cgroupDriver: systemd
# CIDR for pods IP addresses. Needed when using 'kubenet' network plugin and manager-controller is not assigning those.
podCIDR: %s
# Address where kubelet should listen on.
address: %s
# Disable healht port for now, since we don't use it
# TODO check how to use it and re-enable it
healthzPort: 0
# Set up cluster domain. Without this, there is no 'search' field in /etc/resolv.conf in containers, so
# short-names resolution like mysvc.myns.svc does not work.
clusterDomain: cluster.local
# Authenticate clients using CA file
authentication:
  x509:
    clientCAFile: /etc/kubernetes/pki/ca.crt
# Configure cluster DNS IP addresses
clusterDNS:
`, k.podCIDR, k.address)
	// TODO ugly!
	configFiles["/etc/kubernetes/kubelet/kubelet.yaml"] = fmt.Sprintf("%s%s\n", configFiles["/etc/kubernetes/kubelet/kubelet.yaml"], strings.TrimSpace(clusterDNS))

	configFiles["/etc/kubernetes/kubelet/bootstrap-kubeconfig"] = k.bootstrapKubeconfig
	configFiles["/etc/kubernetes/kubelet/pki/ca.crt"] = k.kubernetesCACertificate

	c := container.Container{
		// TODO this is weird. This sets docker as default runtime config
		Runtime: container.RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			// TODO make it configurable?
			Name:  "kubelet",
			Image: k.image,
			// When kubelet runs as a container, it should be privileged, so it can adjust it's OOM settings.
			// Without this, you get following errors:
			// failed to set "/proc/self/oom_score_adj" to "-999": write /proc/self/oom_score_adj: permission denied
			// write /proc/self/oom_score_adj: permission denied
			Privileged: true,
			// Required for detecting node IP address, --node-ip is not enough as kubelet is trying to verify
			// that this IP address is present on the node.
			NetworkMode: "host",
			// Required for adding containers into correct network namespaces
			PidMode: "host",
			Mounts: []types.Mount{
				{
					// Kubelet is using this file to determine what OS it runs on and then reports that to API server
					// If we remove that, kubelet reports as Debian, since by the time of writing, hyperkube images are
					// based on Debian Docker images.
					Source: "/etc/os-release",
					Target: "/etc/os-release",
				},
				{
					// Kubelet will create kubeconfig file for itself from, so it needs to be able to write
					// to /etc/kubernetes, as we want to use default paths when possible in the container.
					// However, on the host, /etc/kubernetes may contain some other files, which shouldn't be
					// visible on kubelet, like etcd pki, so we isolate and expose only ./kubelet subdirectory
					// to the kubelet.
					// TODO make sure if that actually make sense. If kubelet is hijacked, it perhaps has access to entire node
					// anyway
					Source: "/etc/kubernetes/kubelet/",
					Target: "/etc/kubernetes/",
				},
				{
					// Pass docker socket to kubelet container, so it can use it as a container runtime.
					// TODO make it configurable
					// TODO check what happens when Docker daemon gets restarted. Will kubelet be restarted
					// then as well? Should we pass entire /run instead, so new socket gets propagated to container?
					Source: "/run/docker.sock",
					Target: "/var/run/docker.sock",
				},
				{
					// For testing kubenet
					// TODO do we need it?
					Source: "/etc/cni/net.d/",
					Target: "/etc/cni/net.d",
				},
				{
					// TODO do we need it?
					Source: "/opt/cni/bin/",
					Target: "/opt/cni/bin",
				},
				{
					// Required by kubelet when creating Docker containers. rslave borrowed from Rancher.
					// TODO add better explanation
					Source:      "/var/lib/docker/",
					Target:      "/var/lib/docker",
					Propagation: "rslave",
				},
				{
					// Required for kubelet when running Docker containers. Since kubelet mounts stuff there several times
					// mounts should be propagated, hence the "shared". "shared" borrowed from Rancher.
					// TODO add better explanation
					Source:      "/var/lib/kubelet/",
					Target:      "/var/lib/kubelet",
					Propagation: "shared",
				},
				{
					// To persist CNI configuration managed by kubelet. Might be only required with 'kubenet' network plugin.
					// TODO check if this is needed. Maybe explain what is stored there.
					Source: "/var/lib/cni/",
					Target: "/var/lib/cni",
				},
				{
					// For loading kernel modules for kubenet plugin
					Source: "/lib/modules/",
					Target: "/lib/modules",
				},
				{
					// Store pod logs on the host, so they are persistent and also can read by Loki.
					Source: "/var/log/pods/",
					Target: "/var/log/pods",
				},
				{
					// For reading host cgroups, to get stats for docker.service cgroup etc.
					// Without this, following error message occurs:
					// Failed to get system container stats for "/system.slice/kubelet.service": failed to get cgroup stats for "/system.slice/kubelet.service": failed to get container info for "/system.slice/kubelet.service": unknown container "/system.slice/kubelet.service"
					// It seems you can't go any deeper with this mount, otherwise it's not working.
					Source: "/sys/",
					Target: "/sys",
				},
			},
			Args: []string{
				"kubelet",
				// Tell kubelet to use config file
				"--config=/etc/kubernetes/kubelet.yaml",
				// Specify kubeconfig file for kubelet. This enabled API server mode and
				// specifies when kubelet will write kubeconfig file after TLS bootstrapping.
				"--kubeconfig=/etc/kubernetes/kubeconfig",
				// kubeconfig with access token for TLS bootstrapping
				"--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubeconfig",
				// Use 'kubenet' network plugin, as it's the simplest one.
				// TODO allow to use different CNI plugins (just 'cni' to be precise)
				"--network-plugin=kubenet",
				// https://alexbrand.dev/post/why-is-my-kubelet-listening-on-a-random-port-a-closer-look-at-cri-and-the-docker-cri-shim/
				"--redirect-container-streaming=false",
				// --node-ip controls where are exposed nodePort services. Since we want to have them available only on private interface,
				// we specify it equal to address.
				// TODO make it optional/configurable?
				fmt.Sprintf("--node-ip=%s", k.address),
			},
		},
	}

	return &container.HostConfiguredContainer{
		Host:        *k.host,
		ConfigFiles: configFiles,
		Container:   c,
	}
}
