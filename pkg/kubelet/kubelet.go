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
)

// Kubelet represents single kubelet instance.
type Kubelet struct {
	Address             string    `json:"address"`
	Image               string    `json:"image"`
	Host                host.Host `json:"host"`
	BootstrapKubeconfig string    `json:"bootstrapKubeconfig"`
	// TODO we require CA certificate, so it can be referred in bootstrap-kubeconfig. Maybe we should be responsible for creating
	// bootstrap-kubeconfig too then?
	KubernetesCACertificate    string            `json:"kubernetesCACertificate"`
	ClusterDNSIPs              []string          `json:"clusterDNSIPs"`
	Name                       string            `json:"name"`
	Taints                     map[string]string `json:"taints"`
	Labels                     map[string]string `json:"labels"`
	PrivilegedLabels           map[string]string `json:"privilegedLabels"`
	PrivilegedLabelsKubeconfig string            `json:"privilegedLabelsKubeconfig"`
	CgroupDriver               string            `json:"cgroupDriver"`
	NetworkPlugin              string            `json:"networkPlugin"`
	SystemReserved             map[string]string `json:"systemReserved"`
	KubeReserved               map[string]string `json:"kubeReserved"`
	HairpinMode                string            `json:"hairpinMode"`
	VolumePluginDir            string            `json:"volumePluginDir"`

	// Depending on the network plugin, this should be optional, but for now it's required.
	PodCIDR string `json:"podCIDR,omitempty"`
}

// kubelet is a validated, executable version of Kubelet.
type kubelet struct {
	config Kubelet
}

// New validates Kubelet configuration and returns it's usable version.
func (k *Kubelet) New() (container.ResourceInstance, error) {
	// TODO when creating kubelet, also pull pause image using configured Container Runtime to speed up later start of pods?
	if err := k.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate kubelet configuration: %w", err)
	}

	nk := &kubelet{
		config: *k,
	}

	if nk.config.Image == "" {
		nk.config.Image = defaults.KubernetesImage
	}

	return nk, nil
}

// Validate validates kubelet configuration.
//
// TODO better validation should be done here
func (k *Kubelet) Validate() error {
	var errors util.ValidateError

	if k.BootstrapKubeconfig == "" {
		errors = append(errors, fmt.Errorf("bootstrapKubeconfig can't be empty"))
	}

	if k.VolumePluginDir == "" {
		errors = append(errors, fmt.Errorf("volumePluginDir can't be empty"))
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

// config return kubelet configuration file content in YAML format.
func (k *kubelet) configFile() (string, error) {
	config := &kubeletconfig.KubeletConfiguration{
		TypeMeta: v1.TypeMeta{
			Kind:       "KubeletConfiguration",
			APIVersion: kubeletconfig.SchemeGroupVersion.String(),
		},
		// Enables TLS certificate rotation, which is good from security point of view.
		RotateCertificates: true,
		// Request HTTPS server certs from API as well, so kubelet does not generate self-signed certificates.
		ServerTLSBootstrap: true,
		// If Docker is configured to use systemd as a cgroup driver and Docker is used as container
		// runtime, this needs to be set to match Docker.
		// TODO pull that information dynamically based on what container runtime is configured.
		CgroupDriver: k.config.CgroupDriver,
		// Address where kubelet should listen on.
		Address: k.config.Address,
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

		// This defines where should pods cgroups be created, like /kubepods and /kubepods/burstable.
		// Also when specified, it suppresses a lot message about it.
		CgroupRoot: "/",

		// Used for calculating node allocatable resources.
		// If EnforceNodeAllocatable has 'system-reserved' set, those limits will be enforced on cgroup specified
		// with SystemReservedCgroup.
		SystemReserved: k.config.SystemReserved,

		// Used for calculating node allocatable resources.
		// If EnforceNodeAllocatable has 'kube-reserved' set, those limits will be enforced on cgroup specified
		// with KubeReservedCgroup.
		KubeReserved: k.config.KubeReserved,

		ClusterDNS: k.config.ClusterDNSIPs,

		HairpinMode: k.config.HairpinMode,
	}

	if k.config.NetworkPlugin == "kubenet" {
		// CIDR for pods IP addresses. Needed when using 'kubenet' network plugin and manager-controller is not assigning those.
		config.PodCIDR = k.config.PodCIDR
	}

	kubelet, err := yaml.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("serializing to YAML failed: %w", err)
	}

	return string(kubelet), nil
}

func (k *kubelet) configFiles() (map[string]string, error) {
	config, err := k.configFile()
	if err != nil {
		return nil, fmt.Errorf("failed building kubelet configuration: %w", err)
	}

	return map[string]string{
		// kubelet.yaml file is a recommended way to configure the kubelet.
		"/etc/kubernetes/kubelet/kubelet.yaml":         config,
		"/etc/kubernetes/kubelet/bootstrap-kubeconfig": k.config.BootstrapKubeconfig,
		"/etc/kubernetes/kubelet/pki/ca.crt":           k.config.KubernetesCACertificate,
	}, nil
}

// mounts returns kubelet's host mounts.
func (k *kubelet) mounts() []containertypes.Mount {
	return []containertypes.Mount{
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
			// Required when using CNI plugin for networking, as kubelet will verify, that network configuration
			// has been deployed there before creating pods.
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
			// Required by kubelet when creating Docker containers. This is required, when using Docker as container
			// runtime, as cAdvisor, which is integrated into kubelet will try to identify image read-write layer for
			// container when creating a handler for monitoring. This is needed to report disk usage inside the container.
			//
			// It is also needed, as kubelet creates a symlink from Docker container's log file to /var/log/pods.
			Source: "/var/lib/docker/",
			Target: "/var/lib/docker",
		},
		{
			// In there, kubelet persist generated certificates and information about pods. In case of a re-creation of
			// kubelet containers, this information would get lost, so running pods would become orphans, which is not
			// desired.
			//
			// Kubelet also mounts the pods mounts in there, so those directories must be shared with host (where actual
			// Docker containers are created).
			//
			// "shared" propagation is needed, as those pods mounts should be visible for the kubelet as well, otherwise
			// kubelet complains when trying to clean up pods volumes.
			Source:      "/var/lib/kubelet/",
			Target:      "/var/lib/kubelet",
			Propagation: "shared",
		},
		{
			// This is where kubelet stores information about the network configuration on the node when using 'kubenet'
			// as network plugin, so it should be persisted.
			//
			// It is also used for caching network configuration for both 'kubenet' and CNI plugins.
			Source: "/var/lib/cni/",
			Target: "/var/lib/cni",
		},
		{
			// For loading kernel modules for kubenet plugin.
			Source: "/lib/modules/",
			Target: "/lib/modules",
		},
		{
			// In this directory, kubelet creates symlinks to container log files, so this directory should be visible
			// also for other containers. For example for centralised logging, as this is the location, where logging
			// agent expect to find pods logs.
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
		{
			// As kubelet is adding some chains to the iptables, it should share the lock with the host,
			// to avoid races with kube-proxy.
			Source: "/run/xtables.lock",
			Target: "/run/xtables.lock",
		},
		{
			Source: k.config.VolumePluginDir,
			Target: "/usr/libexec/kubernetes/kubelet-plugins/volume/exec",
		},
	}
}

func (k *kubelet) args() []string {
	a := []string{
		"kubelet",
		// Tell kubelet to use config file.
		"--config=/etc/kubernetes/kubelet.yaml",
		// Specify kubeconfig file for kubelet. This enabled API server mode and
		// specifies when kubelet will write kubeconfig file after TLS bootstrapping.
		"--kubeconfig=/var/lib/kubelet/kubeconfig",
		// kubeconfig with access token for TLS bootstrapping.
		"--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubeconfig",
		// Set which network plugin to use.
		fmt.Sprintf("--network-plugin=%s", k.config.NetworkPlugin),
		// https://alexbrand.dev/post/why-is-my-kubelet-listening-on-a-random-port-a-closer-look-at-cri-and-the-docker-cri-shim/
		"--redirect-container-streaming=false",
		// --node-ip controls where are exposed nodePort services. Since we want to have them available only on private interface,
		// we specify it equal to address.
		// TODO make it optional/configurable?
		fmt.Sprintf("--node-ip=%s", k.config.Address),
		// Make sure we register the node with the name specified by the user. This is needed to later on patch the Node object when needed.
		fmt.Sprintf("--hostname-override=%s", k.config.Name),
		// Tell kubelet where to look for CNI binaries. Custom CNI plugins may install their binaries in /opt/cni/host on host filesystem.
		// Also if host filesystem has newer binaries than ones shipped by hyperkube image, those should take precedence.
		//
		// TODO This flag should only be set if Docker is used as container runtime.
		"--cni-bin-dir=/host/opt/cni/bin,/opt/cni/bin",
	}

	if len(k.config.Labels) > 0 {
		a = append(a, fmt.Sprintf("--node-labels=%s", util.JoinSorted(k.config.Labels, "=", ",")))
	}

	if len(k.config.Taints) > 0 {
		a = append(a, fmt.Sprintf("--register-with-taints=%s", util.JoinSorted(k.config.Taints, "=:", ",")))
	}

	return a
}

// ToHostConfiguredContainer takes configured kubelet and converts it to generic HostConfiguredContainer.
func (k *kubelet) ToHostConfiguredContainer() (*container.HostConfiguredContainer, error) {
	configFiles, err := k.configFiles()
	if err != nil {
		return nil, fmt.Errorf("failed building config files map: %w", err)
	}

	c := container.Container{
		// TODO this is weird. This sets docker as default runtime config
		Runtime: container.RuntimeConfig{
			Docker: docker.DefaultConfig(),
		},
		Config: containertypes.ContainerConfig{
			// TODO make it configurable?
			Name:  "kubelet",
			Image: k.config.Image,
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
			Mounts:  k.mounts(),
			Args:    k.args(),
		},
	}

	return &container.HostConfiguredContainer{
		Host:        k.config.Host,
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
	c, err := client.NewClient([]byte(k.config.PrivilegedLabelsKubeconfig))
	if err != nil {
		return fmt.Errorf("failed creating kubernetes client: %w", err)
	}

	return c.LabelNode(k.config.Name, k.config.PrivilegedLabels)
}

// postStartHook defines actions which will be executed after new kubelet instance is created.
func (k *kubelet) postStartHook() *container.Hook {
	f := container.Hook(func() error {
		if len(k.config.PrivilegedLabelsKubeconfig) > 0 {
			if err := k.applyPrivilegedLabels(); err != nil {
				return fmt.Errorf("failed applying privileged labels: %w", err)
			}
		}

		return nil
	})

	return &f
}
