package controlplane

import (
	"fmt"
	"path"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	containertypes "github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/types"
)

// KubeAPIServer represents kube-apiserver container configuration
type KubeAPIServer struct {
	Common                   Common            `json:"common"`
	Host                     host.Host         `json:"host"`
	APIServerCertificate     types.Certificate `json:"apiServerCertificate"`
	APIServerKey             types.PrivateKey  `json:"apiServerKey"`
	ServiceAccountPublicKey  string            `json:"serviceAccountPublicKey"`
	BindAddress              string            `json:"bindAddress"`
	AdvertiseAddress         string            `json:"advertiseAddress"`
	EtcdServers              []string          `json:"etcdServers"`
	ServiceCIDR              string            `json:"serviceCIDR"`
	SecurePort               int               `json:"securePort"`
	FrontProxyCertificate    types.Certificate `json:"frontProxyCertificate"`
	FrontProxyKey            types.PrivateKey  `json:"frontProxyKey"`
	KubeletClientCertificate types.Certificate `json:"kubeletClientCertificate"`
	KubeletClientKey         types.PrivateKey  `json:"kubeletClientKey"`
	EtcdCACertificate        types.Certificate `json:"etcdCACertificate"`
	EtcdClientCertificate    types.Certificate `json:"etcdClientCertificate"`
	EtcdClientKey            types.PrivateKey  `json:"etcdClientKey"`
}

// kubeAPIServer is a validated version of KubeAPIServer
type kubeAPIServer struct {
	common                   Common
	host                     host.Host
	apiServerCertificate     string
	apiServerKey             string
	serviceAccountPublicKey  string
	bindAddress              string
	advertiseAddress         string
	etcdServers              []string
	serviceCIDR              string
	securePort               int
	frontProxyCertificate    string
	frontProxyKey            string
	kubeletClientCertificate string
	kubeletClientKey         string
	etcdCACertificate        string
	etcdClientCertificate    string
	etcdClientKey            string
}

const (
	hostConfigPath      = "/etc/kubernetes/kube-apiserver/pki"
	containerConfigPath = "/etc/kubernetes/pki"
	containerName       = "kube-apiserver"
)

// configFiles returns map of file for kube-apiserver.
func (k *kubeAPIServer) configFiles() map[string]string {
	m := map[string]string{
		"ca.crt":                       string(k.common.KubernetesCACertificate),
		"apiserver.crt":                k.apiServerCertificate,
		"apiserver.key":                k.apiServerKey,
		"service-account.crt":          k.serviceAccountPublicKey,
		"front-proxy-ca.crt":           string(k.common.FrontProxyCACertificate),
		"front-proxy-client.crt":       k.frontProxyCertificate,
		"front-proxy-client.key":       k.frontProxyKey,
		"apiserver-kubelet-client.crt": k.kubeletClientCertificate,
		"apiserver-kubelet-client.key": k.kubeletClientKey,
		"etcd/ca.crt":                  k.etcdCACertificate,
		"apiserver-etcd-client.crt":    k.etcdClientCertificate,
		"apiserver-etcd-client.key":    k.etcdClientKey,
	}

	// Append base path to map.
	for k, v := range m {
		m[path.Join(hostConfigPath, k)] = v
		delete(m, v)
	}

	return m
}

// ToHostConfiguredContainer takes configured values and converts them to generic container configuration
func (k *kubeAPIServer) ToHostConfiguredContainer() (*container.HostConfiguredContainer, error) {
	return &container.HostConfiguredContainer{
		Host:        k.host,
		ConfigFiles: k.configFiles(),
		Container: container.Container{
			// TODO this is weird. This sets docker as default runtime config
			Runtime: container.RuntimeConfig{
				Docker: &docker.Config{},
			},
			Config: containertypes.ContainerConfig{
				Name:  containerName,
				Image: k.common.GetImage(),
				Mounts: []containertypes.Mount{
					{
						Source: hostConfigPath,
						Target: containerConfigPath,
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
					"--requestheader-allowed-names=",
					"--requestheader-extra-headers-prefix=X-Remote-Extra-",
					"--requestheader-group-headers=X-Remote-Group",
					"--requestheader-username-headers=X-Remote-User",
					// Required for communicating with kubelet.
					"--kubelet-client-certificate=/etc/kubernetes/pki/apiserver-kubelet-client.crt",
					"--kubelet-client-key=/etc/kubernetes/pki/apiserver-kubelet-client.key",
					"--kubelet-certificate-authority=/etc/kubernetes/pki/ca.crt",
					// To secure communication to etcd servers.
					"--etcd-cafile=/etc/kubernetes/pki/etcd/ca.crt",
					"--etcd-certfile=/etc/kubernetes/pki/apiserver-etcd-client.crt",
					"--etcd-keyfile=/etc/kubernetes/pki/apiserver-etcd-client.key",
					// Enable additional admission plugins:
					// - NodeRestriction for extra protection against rogue cluster nodes.
					// - PodSecurityPolicy for PSP support.
					"--enable-admission-plugins=NodeRestriction,PodSecurityPolicy",
				},
			},
		},
	}, nil
}

// New validates KubeAPIServer configuration and populates default for some fields, if they are empty
func (k *KubeAPIServer) New() (container.ResourceInstance, error) {
	if err := k.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate Kubernetes API server configuration: %w", err)
	}

	return &kubeAPIServer{
		common:                   k.Common,
		host:                     k.Host,
		apiServerCertificate:     string(k.APIServerCertificate),
		apiServerKey:             string(k.APIServerKey),
		serviceAccountPublicKey:  k.ServiceAccountPublicKey,
		bindAddress:              k.BindAddress,
		advertiseAddress:         k.AdvertiseAddress,
		etcdServers:              k.EtcdServers,
		serviceCIDR:              k.ServiceCIDR,
		securePort:               k.SecurePort,
		frontProxyCertificate:    string(k.FrontProxyCertificate),
		frontProxyKey:            string(k.FrontProxyKey),
		kubeletClientCertificate: string(k.KubeletClientCertificate),
		kubeletClientKey:         string(k.KubeletClientKey),
		etcdCACertificate:        string(k.EtcdCACertificate),
		etcdClientCertificate:    string(k.EtcdClientCertificate),
		etcdClientKey:            string(k.EtcdClientKey),
	}, nil
}

// Validate validates KubeAPIServer struct
//
// TODO add validation of certificates if specified
func (k *KubeAPIServer) Validate() error {
	var errors util.ValidateError

	b, err := yaml.Marshal(k)
	if err != nil {
		return append(errors, fmt.Errorf("failed to validate: %w", err))
	}

	if err := yaml.Unmarshal(b, &k); err != nil {
		return append(errors, fmt.Errorf("validation failed: %w", err))
	}

	if len(k.EtcdServers) == 0 {
		errors = append(errors, fmt.Errorf("at least one etcd server must be defined"))
	}

	if err := k.Host.Validate(); err != nil {
		errors = append(errors, fmt.Errorf("host config validation failed: %w", err))
	}

	return errors.Return()
}
