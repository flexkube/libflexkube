package controlplane

import (
	"fmt"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	containertypes "github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/defaults"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/types"
)

// KubeAPIServer represents kube-apiserver container configuration
type KubeAPIServer struct {
	Image                    string            `json:"image" yaml:"image"`
	Host                     *host.Host        `json:"host" yaml:"host"`
	KubernetesCACertificate  types.Certificate `json:"kubernetesCACertificate" yaml:"kubernetesCACertificate"`
	APIServerCertificate     types.Certificate `json:"apiServerCertificate" yaml:"apiServerCertificate"`
	APIServerKey             types.PrivateKey  `json:"apiServerKey" yaml:"apiServerKey"`
	ServiceAccountPublicKey  string            `json:"serviceAccountPublicKey" yaml:"serviceAccountPublicKey"`
	BindAddress              string            `json:"bindAddress" yaml:"bindAddress"`
	AdvertiseAddress         string            `json:"advertiseAddress" yaml:"advertiseAddress"`
	EtcdServers              []string          `json:"etcdServers" yaml:"etcdServers"`
	ServiceCIDR              string            `json:"serviceCIDR" yaml:"serviceCIDR"`
	SecurePort               int               `json:"securePort" yaml:"securePort"`
	FrontProxyCACertificate  types.Certificate `json:"frontProxyCACertificate" yaml:"frontProxyCACertificate"`
	FrontProxyCertificate    types.Certificate `json:"frontProxyCertificate" yaml:"frontProxyCertificate"`
	FrontProxyKey            types.PrivateKey  `json:"frontProxyKey" yaml:"frontProxyKey"`
	KubeletClientCertificate types.Certificate `json:"kubeletClientCertificate" yaml:"kubeletClientCertificate"`
	KubeletClientKey         types.PrivateKey  `json:"kubeletClientKey" yaml:"kubeletClientKey"`
}

// kubeAPIServer is a validated version of KubeAPIServer
type kubeAPIServer struct {
	image                    string
	host                     host.Host
	kubernetesCACertificate  string
	apiServerCertificate     string
	apiServerKey             string
	serviceAccountPublicKey  string
	bindAddress              string
	advertiseAddress         string
	etcdServers              []string
	serviceCIDR              string
	securePort               int
	frontProxyCACertificate  string
	frontProxyCertificate    string
	frontProxyKey            string
	kubeletClientCertificate string
	kubeletClientKey         string
}

// ToHostConfiguredContainer takes configured values and converts them to generic container configuration
func (k *kubeAPIServer) ToHostConfiguredContainer() *container.HostConfiguredContainer {
	configFiles := make(map[string]string)
	// TODO put all those path in a single place. Perhaps make them configurable with defaults too
	configFiles["/etc/kubernetes/kube-apiserver/pki/ca.crt"] = k.kubernetesCACertificate
	configFiles["/etc/kubernetes/kube-apiserver/pki/apiserver.crt"] = k.apiServerCertificate
	configFiles["/etc/kubernetes/kube-apiserver/pki/apiserver.key"] = k.apiServerKey
	configFiles["/etc/kubernetes/kube-apiserver/pki/service-account.crt"] = k.serviceAccountPublicKey
	configFiles["/etc/kubernetes/kube-apiserver/pki/front-proxy-ca.crt"] = k.frontProxyCACertificate
	configFiles["/etc/kubernetes/kube-apiserver/pki/front-proxy-client.crt"] = k.frontProxyCertificate
	configFiles["/etc/kubernetes/kube-apiserver/pki/front-proxy-client.key"] = k.frontProxyKey
	configFiles["/etc/kubernetes/kube-apiserver/pki/apiserver-kubelet-client.crt"] = k.kubeletClientCertificate
	configFiles["/etc/kubernetes/kube-apiserver/pki/apiserver-kubelet-client.key"] = k.kubeletClientKey

	c := container.Container{
		// TODO this is weird. This sets docker as default runtime config
		Runtime: container.RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: containertypes.ContainerConfig{
			Name:  "kube-apiserver",
			Image: k.image,
			Mounts: []containertypes.Mount{
				{
					Source: "/etc/kubernetes/kube-apiserver/pki/",
					Target: "/etc/kubernetes/pki",
				},
			},
			Ports: []containertypes.PortMap{
				{
					IP:       k.bindAddress,
					Protocol: "tcp",
					Port:     k.securePort,
				},
			},
			Args: []string{
				"kube-apiserver",
				fmt.Sprintf("--etcd-servers=%s", strings.Join(k.etcdServers, ",")),
				"--client-ca-file=/etc/kubernetes/pki/ca.crt",
				"--tls-cert-file=/etc/kubernetes/pki/apiserver.crt",
				"--tls-private-key-file=/etc/kubernetes/pki/apiserver.key",
				// Required for TLS bootstrapping
				"--enable-bootstrap-token-auth=true",
				// Allow user to configure service CIDR, so it does not conflict with host nor pods CIDRs.
				fmt.Sprintf("--service-cluster-ip-range=%s", k.serviceCIDR),
				// To disable access without authentication
				"--insecure-port=0",
				// Since we will run self-hosted K8s, pods like kube-proxy must run as privileged containers, so we must allow them.
				"--allow-privileged=true",
				// Enable RBAC for generic RBAC and Node, so kubelets can use special permissions.
				"--authorization-mode=RBAC,Node",
				// Required to validate service account tokens created by controller manager
				"--service-account-key-file=/etc/kubernetes/pki/service-account.crt",
				// IP address which will be added to the kubernetes.default service endpoint
				fmt.Sprintf("--advertise-address=%s", k.advertiseAddress),
				// For static api-server use non-standard port, so haproxy can use standard one
				fmt.Sprintf("--secure-port=%d", k.securePort),
				// Be a bit more verbose.
				//"--v=2",
				// Prefer to talk to kubelets over InternalIP rather than via Hostname or DNS, to make it more robust
				"--kubelet-preferred-address-types=InternalIP,Hostname,InternalDNS,ExternalDNS,ExternalIP",
				// Required for enabling aggregation layer.
				"--requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt",
				"--proxy-client-key-file=/etc/kubernetes/pki/front-proxy-client.key",
				"--proxy-client-cert-file=/etc/kubernetes/pki/front-proxy-client.crt",
				"--requestheader-allowed-names=\"\"",
				// Required for communicating with kubelet.
				"--kubelet-client-certificate=/etc/kubernetes/pki/apiserver-kubelet-client.crt",
				"--kubelet-client-key=/etc/kubernetes/pki/apiserver-kubelet-client.key",
			},
		},
	}

	return &container.HostConfiguredContainer{
		Host:        k.host,
		ConfigFiles: configFiles,
		Container:   c,
	}
}

// New validates KubeAPIServer configuration and populates default for some fields, if they are empty
func (k *KubeAPIServer) New() (*kubeAPIServer, error) {
	if err := k.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate Kubernetes API server configuration: %w", err)
	}

	nk := &kubeAPIServer{
		image:                    k.Image,
		host:                     *k.Host,
		kubernetesCACertificate:  string(k.KubernetesCACertificate),
		apiServerCertificate:     string(k.APIServerCertificate),
		apiServerKey:             string(k.APIServerKey),
		serviceAccountPublicKey:  k.ServiceAccountPublicKey,
		bindAddress:              k.BindAddress,
		advertiseAddress:         k.AdvertiseAddress,
		etcdServers:              k.EtcdServers,
		serviceCIDR:              k.ServiceCIDR,
		securePort:               k.SecurePort,
		frontProxyCACertificate:  string(k.FrontProxyCACertificate),
		frontProxyCertificate:    string(k.FrontProxyCertificate),
		frontProxyKey:            string(k.FrontProxyKey),
		kubeletClientCertificate: string(k.KubeletClientCertificate),
		kubeletClientKey:         string(k.KubeletClientKey),
	}

	// The only optional parameter
	if nk.image == "" {
		nk.image = defaults.KubernetesImage
	}

	return nk, nil
}

// Validate validates KubeAPIServer struct
//
// TODO add validation of certificates if specified
func (k *KubeAPIServer) Validate() error {
	b, err := yaml.Marshal(k)
	if err != nil {
		return fmt.Errorf("failed to validate: %w", err)
	}

	if err := yaml.Unmarshal(b, &k); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if len(k.EtcdServers) == 0 {
		return fmt.Errorf("at least one etcd server must be defined")
	}

	return nil
}
