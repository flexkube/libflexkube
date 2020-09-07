package controlplane

import (
	"fmt"
	"path"
	"strings"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	containertypes "github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/defaults"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/types"
)

// KubeAPIServer represents kube-apiserver container configuration.
type KubeAPIServer struct {
	// Common stores common information between all controlplane components.
	Common *Common `json:"common,omitempty"`

	// Host defines on which host kube-apiserver container should be created.
	Host *host.Host `json:"host,omitempty"`

	// APIServerCertificate stores X.509 certificate, PEM encoded, which will be
	// used for serving.
	APIServerCertificate types.Certificate `json:"apiServerCertificate"`

	// APIServerKey is a PEM encoded, private key in either PKCS1, PKCS8 or EC format.
	// It must match certificate defined in APIServerCertificate field.
	APIServerKey types.PrivateKey `json:"apiServerKey"`

	// ServiceAccountPublicKey stores PEM encoded public certificate, which will be used
	// to validate service account tokens.
	ServiceAccountPublicKey string `json:"serviceAccountPublicKey"`

	// BindAddress defines IP address where kube-apiserver process should listen for
	// incoming requests.
	BindAddress string `json:"bindAddress"`

	// AdvertiseAddress defines IP address, which should be advertised to
	// kubernetes.default.svc Service on the cluster.
	AdvertiseAddress string `json:"advertiseAddress"`

	// EtcdServers is a list of etcd servers URLs.
	//
	// Example value: '[]string{"https://localhost:2380"}'.
	EtcdServers []string `json:"etcdServers"`

	// ServiceCIDR defines, from which CIDR Service type ClusterIP should get IP addresses
	// assigned. You should make sure, that this CIDR does not collide with any of CIDRs
	// accessible from your cluster nodes.
	//
	// Example value: '10.96.0.0/12'.
	ServiceCIDR string `json:"serviceCIDR"`

	// SecurePort defines TCP port, where kube-apiserver will be listening for incoming
	// requests and which will be advertised to kubernetes.default.svc Service on the cluster.
	//
	// Currently, there is no way to use advertise different port due to kube-apiserver limitations.
	//
	// If you want to mitigate that, you can use APILoadBalancers resource.
	SecurePort int `json:"securePort"`

	// FrontProxyCertificate stores X.509 client certificate, PEM encoded, which will be used by
	// kube-apiserver to talk to extension API server.
	//
	// See https://kubernetes.io/docs/tasks/access-kubernetes-api/configure-aggregation-layer/
	// for more details.
	FrontProxyCertificate types.Certificate `json:"frontProxyCertificate"`

	// FrontProxyKey is a PEM encoded, private key in either PKCS1, PKCS8 or EC format.
	//
	// It must match certificate defined in FrontProxyCertificate field.
	FrontProxyKey types.PrivateKey `json:"frontProxyKey"`

	// KubeletClientCertificate stores X.509 client certificate, PEM encoded, which will be used by
	// kube-apiserver to talk to kubelet process on all nodes, to fetch logs etc.
	KubeletClientCertificate types.Certificate `json:"kubeletClientCertificate"`

	// KubeletClientKey is a PEM encoded, private key in either PKCS1, PKCS8 or EC format.
	//
	// It must match certificate defined in KubeletClientCertificate field.
	KubeletClientKey types.PrivateKey `json:"kubeletClientKey"`

	// EtcdCACertificate stores X.509 CA certificate, PEM encoded, which will be used by
	// kube-apiserver to validate etcd servers certificate.
	EtcdCACertificate types.Certificate `json:"etcdCACertificate"`

	// EtcdClientCertificate stores X.509 client certificate, PEM encoded, which will be used by
	// kube-apiserver to talk to etcd members.
	EtcdClientCertificate types.Certificate `json:"etcdClientCertificate"`

	// EtcdClientKey is a PEM encoded, private key in either PKCS1, PKCS8 or EC format.
	//
	// It must match certificate defined in EtcdClientCertificate field.
	EtcdClientKey types.PrivateKey `json:"etcdClientKey"`
}

// kubeAPIServer is a validated version of KubeAPIServer.
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

	clientCAFile              = "ca.crt"
	tlsCertFile               = "apiserver.crt"
	tlsPrivateKeyFile         = "apiserver.key"
	serviceAccountKeyFile     = "service-account.crt"
	requestheaderClientCAFile = "front-proxy-ca.crt"
	proxyClientCertFile       = "front-proxy-client.crt"
	proxyClientKeyFile        = "front-proxy-client.key"
	kubeletClientCertificate  = "apiserver-kubelet-client.crt"
	kubeletClientKey          = "apiserver-kubelet-client.key"
	etcdCAFile                = "etcd/ca.crt"
	etcdCertificate           = "apiserver-etcd-client.crt"
	etcdKeyfile               = "apiserver-etcd-client.key"
)

// configFiles returns map of file for kube-apiserver.
func (k *kubeAPIServer) configFiles() map[string]string {
	m := map[string]string{
		clientCAFile:              string(k.common.KubernetesCACertificate),
		tlsCertFile:               k.apiServerCertificate,
		tlsPrivateKeyFile:         k.apiServerKey,
		serviceAccountKeyFile:     k.serviceAccountPublicKey,
		requestheaderClientCAFile: string(k.common.FrontProxyCACertificate),
		proxyClientCertFile:       k.frontProxyCertificate,
		proxyClientKeyFile:        k.frontProxyKey,
		kubeletClientCertificate:  k.kubeletClientCertificate,
		kubeletClientKey:          k.kubeletClientKey,
		etcdCAFile:                k.etcdCACertificate,
		etcdCertificate:           k.etcdClientCertificate,
		etcdKeyfile:               k.etcdClientKey,
	}

	r := map[string]string{}

	// Append base path to map.
	for k, v := range m {
		r[path.Join(hostConfigPath, k)] = v
	}

	return r
}

// args returns kube-apiserver set of flags.
func (k *kubeAPIServer) args() []string {
	return []string{
		"kube-apiserver",
		fmt.Sprintf("--etcd-servers=%s", strings.Join(k.etcdServers, ",")),
		fmt.Sprintf("--client-ca-file=%s", path.Join(containerConfigPath, clientCAFile)),
		fmt.Sprintf("--tls-cert-file=%s", path.Join(containerConfigPath, tlsCertFile)),
		fmt.Sprintf("--tls-private-key-file=%s", path.Join(containerConfigPath, tlsPrivateKeyFile)),
		// Required for TLS bootstrapping.
		"--enable-bootstrap-token-auth=true",
		// Allow user to configure service CIDR, so it does not conflict with host nor pods CIDRs.
		fmt.Sprintf("--service-cluster-ip-range=%s", k.serviceCIDR),
		// To disable access without authentication.
		"--insecure-port=0",
		// Since we will run self-hosted K8s, pods like kube-proxy must run as privileged containers, so we must allow them.
		"--allow-privileged=true",
		// Enable RBAC for generic RBAC and Node, so kubelets can use special permissions.
		"--authorization-mode=RBAC,Node",
		// Required to validate service account tokens created by controller manager.
		fmt.Sprintf("--service-account-key-file=%s", path.Join(containerConfigPath, serviceAccountKeyFile)),
		// IP address which will be added to the kubernetes.default service endpoint.
		fmt.Sprintf("--advertise-address=%s", k.advertiseAddress),
		// For static api-server use non-standard port, so haproxy can use standard one.
		fmt.Sprintf("--secure-port=%d", k.securePort),
		// Prefer to talk to kubelets over InternalIP rather than via Hostname or DNS, to make it more robust.
		"--kubelet-preferred-address-types=InternalIP,Hostname,InternalDNS,ExternalDNS,ExternalIP",
		// Required for enabling aggregation layer.
		fmt.Sprintf("--requestheader-client-ca-file=%s", path.Join(containerConfigPath, requestheaderClientCAFile)),
		fmt.Sprintf("--proxy-client-cert-file=%s", path.Join(containerConfigPath, proxyClientCertFile)),
		fmt.Sprintf("--proxy-client-key-file=%s", path.Join(containerConfigPath, proxyClientKeyFile)),
		"--requestheader-allowed-names=",
		"--requestheader-extra-headers-prefix=X-Remote-Extra-",
		"--requestheader-group-headers=X-Remote-Group",
		"--requestheader-username-headers=X-Remote-User",
		// Required for communicating with kubelet.
		fmt.Sprintf("--kubelet-client-certificate=%s", path.Join(containerConfigPath, kubeletClientCertificate)),
		fmt.Sprintf("--kubelet-client-key=%s", path.Join(containerConfigPath, kubeletClientKey)),
		fmt.Sprintf("--kubelet-certificate-authority=%s", path.Join(containerConfigPath, clientCAFile)),
		// To secure communication to etcd servers.
		fmt.Sprintf("--etcd-cafile=%s", path.Join(containerConfigPath, etcdCAFile)),
		fmt.Sprintf("--etcd-certfile=%s", path.Join(containerConfigPath, etcdCertificate)),
		fmt.Sprintf("--etcd-keyfile=%s", path.Join(containerConfigPath, etcdKeyfile)),
		// Enable additional admission plugins:
		// - NodeRestriction for extra protection against rogue cluster nodes.
		// - PodSecurityPolicy for PSP support.
		"--enable-admission-plugins=NodeRestriction,PodSecurityPolicy",
		// To limit memory consumption of bootstrap controlplane, limit it to 512 MB.
		"--target-ram-mb=512",
		// Use SO_REUSEPORT, so multiple instances can run on the same controller for smooth upgrades.
		"--permit-port-sharing=true",
	}
}

// ToHostConfiguredContainer takes configured values and converts them to generic container configuration.
func (k *kubeAPIServer) ToHostConfiguredContainer() (*container.HostConfiguredContainer, error) {
	return &container.HostConfiguredContainer{
		Host:        k.host,
		ConfigFiles: k.configFiles(),
		Container: container.Container{
			// TODO: This is weird. This sets docker as default runtime config.
			Runtime: container.RuntimeConfig{
				Docker: docker.DefaultConfig(),
			},
			Config: containertypes.ContainerConfig{
				Name:        containerName,
				Image:       util.PickString(k.common.Image, defaults.KubeAPIServerImage),
				NetworkMode: "host",
				Mounts: []containertypes.Mount{
					{
						Source: hostConfigPath,
						Target: containerConfigPath,
					},
				},
				Args: k.args(),
			},
		},
	}, nil
}

// New validates KubeAPIServer configuration and populates default for some fields, if they are empty.
func (k *KubeAPIServer) New() (container.ResourceInstance, error) {
	if k.Common == nil {
		k.Common = &Common{}
	}

	if k.Host == nil {
		k.Host = &host.Host{}
	}

	if err := k.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate Kubernetes API server configuration: %w", err)
	}

	return &kubeAPIServer{
		common:                   *k.Common,
		host:                     *k.Host,
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

// Validate validates KubeAPIServer struct.
//
// TODO: Add validation of certificates if specified.
func (k *KubeAPIServer) Validate() error {
	var errors util.ValidateError

	v := validator{
		Common: k.Common,
		Host:   k.Host,
		YAML:   k,
	}

	if err := v.validate(false); err != nil {
		errors = append(errors, err)
	}

	if len(k.EtcdServers) == 0 {
		errors = append(errors, fmt.Errorf("at least one etcd server must be defined"))
	}

	return errors.Return()
}
