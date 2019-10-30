package controlplane

import (
	"fmt"
	"strings"

	"github.com/invidian/libflexkube/pkg/container"
	"github.com/invidian/libflexkube/pkg/container/runtime/docker"
	"github.com/invidian/libflexkube/pkg/container/types"
	"github.com/invidian/libflexkube/pkg/defaults"
	"github.com/invidian/libflexkube/pkg/host"
)

type KubeAPIServer struct {
	Image                   string     `json:"image,omitempty" yaml:"image,omitempty"`
	Host                    *host.Host `json:"host,omitempty" yaml:"host,omitempty"`
	KubernetesCACertificate string     `json:"kubernetesCACertificate,omitempty" yaml:"kubernetesCACertificate,omitempty"`
	APIServerCertificate    string     `json:"apiServerCertificate,omitempty" yaml:"apiServerCertificate,omitempty"`
	APIServerKey            string     `json:"apiServerKey,omitempty" yaml:"apiServerKey,omitempty"`
	ServiceAccountPublicKey string     `json:"serviceAccountPublicKey,omitempty" yaml:"serviceAccountPublicKey,omitempty"`
	Address                 string     `json:"address,omitempty" yaml:"address,omitempty"`
	EtcdServers             []string   `json:"etcdServers,omitempty" yaml:"etcdServers,omitempty"`
}

type kubeAPIServer struct {
	image                   string
	host                    host.Host
	kubernetesCACertificate string
	apiServerCertificate    string
	apiServerKey            string
	serviceAccountPublicKey string
	address                 string
	etcdServers             []string
}

func (k *kubeAPIServer) ToHostConfiguredContainer() *container.HostConfiguredContainer {
	configFiles := make(map[string]string)
	// TODO put all those path in a single place. Perhaps make them configurable with defaults too
	configFiles["/etc/kubernetes/pki/ca.crt"] = k.kubernetesCACertificate
	configFiles["/etc/kubernetes/pki/apiserver.crt"] = k.apiServerCertificate
	configFiles["/etc/kubernetes/pki/apiserver.key"] = k.apiServerKey
	configFiles["/etc/kubernetes/pki/service-account.crt"] = k.serviceAccountPublicKey

	c := container.Container{
		// TODO this is weird. This sets docker as default runtime config
		Runtime: container.RuntimeConfig{
			Docker: &docker.ClientConfig{},
		},
		Config: types.ContainerConfig{
			Name:       "kube-apiserver",
			Image:      k.image,
			Entrypoint: []string{"/hyperkube"},
			Mounts: []types.Mount{
				types.Mount{
					Source: "/etc/kubernetes/pki/ca.crt",
					Target: "/etc/kubernetes/pki/ca.crt",
				},
				types.Mount{
					Source: "/etc/kubernetes/pki/apiserver.crt",
					Target: "/etc/kubernetes/pki/apiserver.crt",
				},
				types.Mount{
					Source: "/etc/kubernetes/pki/apiserver.key",
					Target: "/etc/kubernetes/pki/apiserver.key",
				},
				types.Mount{
					Source: "/etc/kubernetes/pki/service-account.crt",
					Target: "/etc/kubernetes/pki/service-account.crt",
				},
			},
			Ports: []types.PortMap{
				types.PortMap{
					IP:       k.address,
					Protocol: "tcp",
					// TODO make port configurable
					Port: 8443,
				},
			},
			Args: []string{
				"kube-apiserver",
				fmt.Sprintf("--etcd-servers=%s", strings.Join(k.etcdServers[:], ",")),
				"--client-ca-file=/etc/kubernetes/pki/ca.crt",
				"--tls-cert-file=/etc/kubernetes/pki/apiserver.crt",
				"--tls-private-key-file=/etc/kubernetes/pki/apiserver.key",
				// Required for TLS bootstrapping
				"--enable-bootstrap-token-auth=true",
				// Override default service cluster IP, as it conflicts with host CIDR.
				// TODO make it configurable
				"--service-cluster-ip-range=11.0.0.0/24",
				// To disable access without authentication
				"--insecure-port=0",
				// Since we will run self-hosted K8s, pods like kube-proxy must run as privileged containers, so we must allow them.
				"--allow-privileged=true",
				// Enable RBAC for generic RBAC and Node, so kubelets can use special permissions.
				"--authorization-mode=RBAC,Node",
				// Required to validate service account tokens created by controller manager
				"--service-account-key-file=/etc/kubernetes/pki/service-account.crt",
				// IP address which will be added to the kubernetes.default service endpoint
				fmt.Sprintf("--advertise-address=%s", k.address),
				// For static api-server use non-standard port, so haproxy can use standard one
				"--secure-port=8443",
				// Be a bit more verbose.
				"--v=2",
				// Prefer to talk to kubelets over InternalIP rather than via Hostname or DNS, to make it more robust
				"--kubelet-preferred-address-types=InternalIP,Hostname,InternalDNS,ExternalDNS,ExternalIP",
			},
		},
	}

	return &container.HostConfiguredContainer{
		Host:        k.host,
		ConfigFiles: configFiles,
		Container:   c,
	}
}

func (k *KubeAPIServer) New() (*kubeAPIServer, error) {
	if err := k.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate Kubernetes API server configuration: %w", err)
	}

	nk := &kubeAPIServer{
		image:                   k.Image,
		host:                    *k.Host,
		kubernetesCACertificate: k.KubernetesCACertificate,
		apiServerCertificate:    k.APIServerCertificate,
		apiServerKey:            k.APIServerKey,
		serviceAccountPublicKey: k.ServiceAccountPublicKey,
		address:                 k.Address,
		etcdServers:             k.EtcdServers,
	}

	// The only optional parameter
	if nk.image == "" {
		nk.image = defaults.KubernetesImage
	}

	return nk, nil
}

// TODO add validation of certificates if specified
func (k *KubeAPIServer) Validate() error {
	if k.KubernetesCACertificate == "" {
		return fmt.Errorf("KubernetesCACertificate is empty")
	}
	if k.APIServerCertificate == "" {
		return fmt.Errorf("ApiServerCertificate is empty")
	}
	if k.APIServerKey == "" {
		return fmt.Errorf("ApiServerKey is empty")
	}
	if k.ServiceAccountPublicKey == "" {
		return fmt.Errorf("ServiceAccountPublicKey is empty")
	}
	if k.Address == "" {
		return fmt.Errorf("Address is empty")
	}
	if len(k.EtcdServers) == 0 {
		return fmt.Errorf("At least one etcd server must be defined")
	}

	return nil
}
