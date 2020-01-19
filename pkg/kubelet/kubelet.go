package kubelet

import (
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeletconfig "k8s.io/kubelet/config/v1beta1"
	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	containertypes "github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/defaults"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
	"github.com/flexkube/libflexkube/pkg/types"
)

// Kubelet represents single kubelet instance.
type Kubelet struct {
	Address             string    `json:"address" yaml:"address"`
	Image               string    `json:"image" yaml:"image"`
	Host                host.Host `json:"host" yaml:"host"`
	BootstrapKubeconfig string    `json:"bootstrapKubeconfig" yaml:"bootstrapKubeconfig"`
	// TODO we require CA certificate, so it can be referred in bootstrap-kubeconfig. Maybe we should be responsible for creating
	// bootstrap-kubeconfig too then?
	KubernetesCACertificate    string            `json:"kubernetesCACertificate" yaml:"kubernetesCACertificate"`
	ClusterDNSIPs              []string          `json:"clusterDNSIPs" yaml:"clusterDNSIPs"`
	Name                       string            `json:"name" yaml:"name"`
	Taints                     map[string]string `json:"taints" yaml:"taints"`
	Labels                     map[string]string `json:"labels" yaml:"labels"`
	PrivilegedLabels           map[string]string `json:"privilegedLabels" yaml:"privilegedLabels"`
	PrivilegedLabelsKubeconfig string            `json:"privilegedLabelsKubeconfig" yaml:"privilegedLabelsKubeconfig"`
	CgroupDriver               string            `json:"cgroupDriver" yaml:"cgroupDriver"`
	NetworkPlugin              string            `json:"networkPlugin" yaml:"networkPlugin"`

	// Depending on the network plugin, this should be optional, but for now it's required.
	PodCIDR string `json:"podCIDR,omitempty" yaml:"podCIDR,omitempty"`
}

// kubelet is a validated, executable version of Kubelet.
type kubelet struct {
	address                    string
	image                      string
	host                       host.Host
	bootstrapKubeconfig        string
	kubernetesCACertificate    string
	clusterDNSIPs              []string
	podCIDR                    string
	name                       string
	taints                     map[string]string
	labels                     map[string]string
	privilegedLabels           map[string]string
	privilegedLabelsKubeconfig string
	cgroupDriver               string
	networkPlugin              string
}

// New validates Kubelet configuration and returns it's usable version.
func (k *Kubelet) New() (container.ResourceInstance, error) {
	// TODO when creating kubelet, also pull pause image using configured Container Runtime to speed up later start of pods?
	if err := k.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate kubelet configuration: %w", err)
	}

	nk := &kubelet{
		image:                      k.Image,
		address:                    k.Address,
		host:                       k.Host,
		bootstrapKubeconfig:        k.BootstrapKubeconfig,
		kubernetesCACertificate:    k.KubernetesCACertificate,
		clusterDNSIPs:              k.ClusterDNSIPs,
		podCIDR:                    k.PodCIDR,
		name:                       k.Name,
		taints:                     k.Taints,
		labels:                     k.Labels,
		privilegedLabels:           k.PrivilegedLabels,
		privilegedLabelsKubeconfig: k.PrivilegedLabelsKubeconfig,
		cgroupDriver:               k.CgroupDriver,
		networkPlugin:              k.NetworkPlugin,
	}

	if nk.image == "" {
		nk.image = defaults.KubernetesImage
	}

	return nk, nil
}

// Validate validates kubelet configuration.
//
// TODO better validation should be done here
func (k *Kubelet) Validate() error {
	var errors types.ValidateError

	if k.BootstrapKubeconfig == "" {
		errors = append(errors, fmt.Errorf("bootstrapKubeconfig can't be empty"))
	}

	if len(k.PrivilegedLabels) > 0 && k.PrivilegedLabelsKubeconfig == "" {
		errors = append(errors, fmt.Errorf("privilegedLabels requested, but privilegedLabelsKubeconfig is empty"))
	}

	if k.PrivilegedLabelsKubeconfig != "" && len(k.PrivilegedLabels) == 0 {
		errors = append(errors, fmt.Errorf("privilegedLabelsKubeconfig specified, but no privilegedLabels requested"))
	}

	switch k.NetworkPlugin {
	case "cni":
		if k.PodCIDR != "" {
			errors = append(errors, fmt.Errorf("podCIDR has no effect when using 'cni' network plugin"))
		}
	case "kubenet":
		if k.PodCIDR == "" {
			errors = append(errors, fmt.Errorf("podCIDR must be set when using 'kubenet' network plugin"))
		}
	default:
		errors = append(errors, fmt.Errorf("networkPlugin must be either 'cni' or 'kubenet'"))
	}

	if err := k.Host.Validate(); err != nil {
		errors = append(errors, fmt.Errorf("host validation failed: %w", err))
	}

	return errors.Return()
}

// ToHostConfiguredContainer takes configured kubelet and converts it to generic HostConfiguredContainer.
func (k *kubelet) ToHostConfiguredContainer() (*container.HostConfiguredContainer, error) {
	configFiles := make(map[string]string)

	config := &kubeletconfig.KubeletConfiguration{
		TypeMeta: v1.TypeMeta{
			Kind:       "KubeletConfiguration",
			APIVersion: kubeletconfig.SchemeGroupVersion.String(),
		},
		// Enables TLS certificate rotation, which is good from security point of view.
		RotateCertificates: true,
		// Request HTTPS server certs from API as well, so kubelet does not generate self-signed certificates.
		ServerTLSBootstrap: true,
		// To address: "--cgroups-per-qos enabled, but --cgroup-root was not specified.  defaulting to /"
		// This disables QoS based cgroup hierarchy, which is important from resource management perspective.
		CgroupsPerQOS: &[]bool{false}[0],
		// When cgroupsPerQOS is false, enforceNodeAllocatable needs to be set explicitly to empty.
		// TODO This will be removed by yaml.Marshal, so we add it manually later.
		EnforceNodeAllocatable: []string{},
		// If Docker is configured to use systemd as a cgroup driver and Docker is used as container
		// runtime, this needs to be set to match Docker.
		// TODO pull that information dynamically based on what container runtime is configured.
		CgroupDriver: k.cgroupDriver,
		// Address where kubelet should listen on.
		Address: k.address,
		// Disable healht port for now, since we don't use it.
		// TODO check how to use it and re-enable it.
		HealthzPort: &[]int32{0}[0],
		// Set up cluster domain. Without this, there is no 'search' field in /etc/resolv.conf in containers, so
		// short-names resolution like mysvc.myns.svc does not work.
		ClusterDomain: "cluster.local",
		// Authenticate clients using CA file.
		Authentication: kubeletconfig.KubeletAuthentication{
			X509: kubeletconfig.KubeletX509Authentication{
				ClientCAFile: "/etc/kubernetes/pki/ca.crt",
			},
		},
		ClusterDNS: k.clusterDNSIPs,
	}

	if k.networkPlugin == "kubenet" {
		// CIDR for pods IP addresses. Needed when using 'kubenet' network plugin and manager-controller is not assigning those.
		config.PodCIDR = k.podCIDR
	}

	kubelet, err := yaml.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed serializing kubelet configuration: %w", err)
	}

	// kubelet.yaml file is a recommended way to configure the kubelet.
	configFiles["/etc/kubernetes/kubelet/kubelet.yaml"] = string(kubelet)

	// When cgroupsPerQOS is false, enforceNodeAllocatable needs to be set explicitly to empty.
	// TODO Figure out how to put that in the struct up there, as when doing yaml.Marshal, it removes empty slice.
	configFiles["/etc/kubernetes/kubelet/kubelet.yaml"] = fmt.Sprintf("%senforceNodeAllocatable: []\n", configFiles["/etc/kubernetes/kubelet/kubelet.yaml"])

	configFiles["/etc/kubernetes/kubelet/bootstrap-kubeconfig"] = k.bootstrapKubeconfig
	configFiles["/etc/kubernetes/kubelet/pki/ca.crt"] = k.kubernetesCACertificate

	c := container.Container{
		// TODO this is weird. This sets docker as default runtime config
		Runtime: container.RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: containertypes.ContainerConfig{
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
			// Required for adding containers into correct network namespaces.
			PidMode: "host",
			Mounts: []containertypes.Mount{
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
					// For testing kubenet.
					// TODO do we need it?
					Source: "/etc/cni/net.d/",
					Target: "/etc/cni/net.d",
				},
				{
					// Mount host CNI binaries into the kubelet, so it can access it.
					// Hyperkube image already ships some CNI binaries, so we shouldn't shadow them in the kubelet.
					// Other CNI plugins may install CNI binaries in this directory on host, so kubelet should have
					// access to both locations.
					Source: "/opt/cni/bin/",
					Target: "/host/opt/cni/bin",
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
					// For loading kernel modules for kubenet plugin.
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
				// Tell kubelet to use config file.
				"--config=/etc/kubernetes/kubelet.yaml",
				// Specify kubeconfig file for kubelet. This enabled API server mode and
				// specifies when kubelet will write kubeconfig file after TLS bootstrapping.
				"--kubeconfig=/etc/kubernetes/kubeconfig",
				// kubeconfig with access token for TLS bootstrapping.
				"--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubeconfig",
				// Use 'kubenet' network plugin, as it's the simplest one.
				// TODO allow to use different CNI plugins (just 'cni' to be precise)
				fmt.Sprintf("--network-plugin=%s", k.networkPlugin),
				// https://alexbrand.dev/post/why-is-my-kubelet-listening-on-a-random-port-a-closer-look-at-cri-and-the-docker-cri-shim/
				"--redirect-container-streaming=false",
				// --node-ip controls where are exposed nodePort services. Since we want to have them available only on private interface,
				// we specify it equal to address.
				// TODO make it optional/configurable?
				fmt.Sprintf("--node-ip=%s", k.address),
				// Make sure we register the node with the name specified by the user. This is needed to later on patch the Node object when needed.
				fmt.Sprintf("--hostname-override=%s", k.name),
				// Tell kubelet where to look for CNI binaries. Custom CNI plugins may install their binaries in /opt/cni/host on host filesystem.
				// Also if host filesystem has newer binaries than ones shipped by hyperkube image, those should take precedence.
				//
				// TODO This flag should only be set if Docker is used as container runtime.
				"--cni-bin-dir=/host/opt/cni/bin,/opt/cni/bin",
			},
		},
	}

	if len(k.labels) > 0 {
		c.Config.Args = append(c.Config.Args, fmt.Sprintf("--node-labels=%s", util.JoinSorted(k.labels, "=", ",")))
	}

	if len(k.taints) > 0 {
		c.Config.Args = append(c.Config.Args, fmt.Sprintf("--register-with-taints=%s", util.JoinSorted(k.taints, "=:", ",")))
	}

	return &container.HostConfiguredContainer{
		Host:        k.host,
		ConfigFiles: configFiles,
		Container:   c,
		Hooks:       k.getHooks(),
	}, nil
}

// getHooks returns HostConfiguredContainer hooks associated with kubelet.
func (k *kubelet) getHooks() *container.Hooks {
	return &container.Hooks{
		PostStart: k.postStartHook(),
	}
}

// applyPrivilegedLabels adds privileged labels to kubelet object using Kubernetes API.
func (k *kubelet) applyPrivilegedLabels() error {
	c, err := client.NewClient([]byte(k.privilegedLabelsKubeconfig))
	if err != nil {
		return fmt.Errorf("failed creating kubernetes client: %w", err)
	}

	return c.LabelNode(k.name, k.privilegedLabels)
}

// postStartHook defines actions which will be executed after new kubelet instance is created.
func (k *kubelet) postStartHook() *container.Hook {
	f := container.Hook(func() error {
		if len(k.privilegedLabelsKubeconfig) > 0 {
			if err := k.applyPrivilegedLabels(); err != nil {
				return fmt.Errorf("failed applying privileged labels: %w", err)
			}
		}

		return nil
	})

	return &f
}
