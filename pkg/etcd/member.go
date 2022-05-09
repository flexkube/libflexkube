package etcd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"strings"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	containertypes "github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/pki"
	"github.com/flexkube/libflexkube/pkg/types"
)

// MemberConfig represents single etcd member.
type MemberConfig struct {
	// Name defines the name of the etcd member. It is used for --name flag.
	//
	// Example values: etcd01, infra2, member3
	//
	// This field is optional if used with Cluster struct.
	Name string `json:"name,omitempty"`

	// Image is a Docker image with tag to use for member container.
	//
	// Example values: 'quay.io/coreos/etcd:v3.4.9'
	//
	// This field is optional if user together with Cluster struct.
	Image string `json:"image,omitempty"`

	// Host describes on which machine member container should be created.
	//
	// This field is required.
	Host host.Host `json:"host,omitempty"`

	// CACertificate is a etcd CA X.509 certificate used to verify peers and client
	// certificates. It is used for --peer-trusted-ca-file and --trusted-ca-file flags.
	//
	// This certificate can be generated using pki.PKI struct.
	//
	// This field is optional, if used together with Cluster struct.
	CACertificate string `json:"caCertificate,omitempty"`

	// PeerCertificate is a X.509 certificate used to communicate with other cluster
	// members. Should be signed by CACertificate. It is used for --peer-cert-file flag.
	//
	// This certificate can be generated using pki.PKI struct.
	//
	// This field is optional, if used together with Cluster struct and PKI integration.
	PeerCertificate string `json:"peerCertificate,omitempty"`

	// PeerKey is a private key for PeerCertificate. Must be defined in either
	// PKCS8, PKCS1 or EC formats, PEM encoded. It is used for --peer-key-file flag.
	//
	// This private key can be generated using pki.PKI struct.
	//
	// This field is optional, if used together with Cluster struct and PKI integration.
	PeerKey string `json:"peerKey,omitempty"`

	// PeerAddress is an address, where member will listen and which will be
	// advertised to the cluster. It is used for --listen-peer-urls and
	// --initial-advertise-peer-urls flags.
	//
	// Example value: 192.168.10.10
	PeerAddress string `json:"peerAddress,omitempty"`

	// InitialCluster defines initial list of members for the cluster. It is used for
	// --initial-cluster flag.
	//
	// Example value: 'infra0=http://10.0.1.10:2380,infra1=http://10.0.1.11:2380'.
	//
	// This field is optional, if used together with Cluster struct.
	InitialCluster string `json:"initialCluster,omitempty"`

	// PeerCertAllowedCN defines allowed CommonName of the client certificate
	// for peer communication. Can be used when single client certificate is used
	// for all members of the cluster.
	//
	// Is is used for --peer-cert-allowed-cn flag.
	//
	// Example value: 'member'.
	//
	// This field is optional.
	PeerCertAllowedCN string `json:"peerCertAllowedCN,omitempty"`

	// ServerCertificate is a X.509 certificate used to communicate with other cluster
	// members. Should be signed by CACertificate. It is used for --peer-cert-file flag.
	//
	//
	// This certificate can be generated using pki.PKI struct.
	//
	// This field is optional, if used together with Cluster struct and PKI integration.
	ServerCertificate string `json:"serverCertificate,omitempty"`

	// Serverkey is a private key for ServerCertificate. Must be defined in either
	// PKCS8, PKCS1 or EC formats, PEM encoded. It is used for --peer-key-file flag.
	//
	// This private key can be generated using pki.PKI struct.
	//
	// This field is optional, if used together with Cluster struct and PKI integration.
	ServerKey string `json:"serverKey,omitempty"`

	// ServerAddress is an address, where member will listen and which will be
	// advertised to the clients. It is used for --listen-client-urls and
	// --advertise-client-urls flags.
	//
	// Example value: 192.168.10.10
	ServerAddress string `json:"serverAddress,omitempty"`

	// NewCluster controls if member should be created as part of new cluster or as part
	// of already initialized cluster.
	//
	// If set to true, --initial-cluster-token flag will be used when creating the container,
	// otherwise --initial-cluster-state=existing flag will be used.
	//
	// This field is optional, if used together with Cluster struct.
	NewCluster bool `json:"newCluster,omitempty"`

	// ExtraMounts defines extra mounts from host filesystem, which should be added to kubelet
	// containers. It will be used unless kubelet instance define it's own extra mounts.
	ExtraMounts []containertypes.Mount `json:"extraMounts,omitempty"`
}

// Member represents functionality provided by validated MemberConfig.
type Member interface {
	container.ResourceInstance

	peerAddress() string
	add(cli etcdClient) error
	forwardEndpoints(endpoints []string) ([]string, error)
	getEtcdClient(endpoints []string) (etcdClient, error)
}

// member is a validated, executable version of MemberConfig.
type member struct {
	config *MemberConfig
}

func (m *member) configFiles() map[string]string {
	return map[string]string{
		"/etc/kubernetes/etcd/ca.crt":     m.config.CACertificate,
		"/etc/kubernetes/etcd/peer.crt":   m.config.PeerCertificate,
		"/etc/kubernetes/etcd/peer.key":   m.config.PeerKey,
		"/etc/kubernetes/etcd/server.crt": m.config.ServerCertificate,
		"/etc/kubernetes/etcd/server.key": m.config.ServerKey,
	}
}

// args returns flags which will be set to the container.
func (m *member) args() []string {
	authToken := strings.Join([]string{
		"jwt",
		"pub-key=/etc/kubernetes/pki/etcd/peer.crt",
		"priv-key=/etc/kubernetes/pki/etcd/peer.key",
		"sign-method=RS512",
		"ttl=10m",
	}, ",")

	flags := []string{
		// TODO Add descriptions explaining why we need each line.
		// Default value 'capnslog' for logger is deprecated and prints warning now.
		"--logger=zap", // Available only from 3.4.x
		// Since we are in container, listen on all interfaces.
		fmt.Sprintf("--listen-client-urls=https://%s:2379", m.config.ServerAddress),
		fmt.Sprintf("--listen-peer-urls=https://%s:2380", m.config.PeerAddress),
		fmt.Sprintf("--advertise-client-urls=https://%s:2379", m.config.ServerAddress),
		fmt.Sprintf("--initial-advertise-peer-urls=https://%s:2380", m.config.PeerAddress),
		fmt.Sprintf("--initial-cluster=%s", m.config.InitialCluster),
		fmt.Sprintf("--name=%s", m.config.Name),
		"--peer-trusted-ca-file=/etc/kubernetes/pki/etcd/ca.crt",
		"--peer-cert-file=/etc/kubernetes/pki/etcd/peer.crt",
		"--peer-key-file=/etc/kubernetes/pki/etcd/peer.key",
		"--peer-client-cert-auth",
		"--trusted-ca-file=/etc/kubernetes/pki/etcd/ca.crt",
		"--cert-file=/etc/kubernetes/pki/etcd/server.crt",
		"--key-file=/etc/kubernetes/pki/etcd/server.key",
		fmt.Sprintf("--data-dir=/%s.etcd", m.config.Name),
		// To get rid of warning with default configuration.
		// ttl parameter support has been added in 3.4.x.
		fmt.Sprintf("--auth-token=%s", authToken),
		// This is set by typhoon, seems like extra safety knob.
		"--strict-reconfig-check",
		// TODO: Enable metrics.
		// Enable TLS authentication with certificate CN field.
		// See https://github.com/etcd-io/etcd/blob/master/Documentation/op-guide/authentication.md#using-tls-common-name
		// for more details.
		"--client-cert-auth=true",
	}

	if m.config.PeerCertAllowedCN != "" {
		flags = append(flags, fmt.Sprintf("--peer-cert-allowed-cn=%s", m.config.PeerCertAllowedCN))
	}

	return flags
}

// ToHostConfiguredContainer takes configured member and converts it to generic HostConfiguredContainer.
func (m *member) ToHostConfiguredContainer() (*container.HostConfiguredContainer, error) {
	memberContainer := container.Container{
		// TODO: This is weird. This sets docker as default runtime config.
		Runtime: container.RuntimeConfig{
			Docker: docker.DefaultConfig(),
		},
		Config: containertypes.ContainerConfig{
			Name:       fmt.Sprintf("etcd-%s", m.config.Name),
			Image:      m.config.Image,
			Entrypoint: []string{"/usr/local/bin/etcd"},
			Mounts: append(
				[]containertypes.Mount{
					{
						// TODO: Between /var/lib/etcd and data dir we should probably put cluster name, to group them.
						// TODO: Make data dir configurable.
						Source: fmt.Sprintf("/var/lib/etcd/%s.etcd/", m.config.Name),
						Target: fmt.Sprintf("/%s.etcd", m.config.Name),
					},
					{
						Source: "/etc/kubernetes/etcd/",
						Target: "/etc/kubernetes/pki/etcd",
					},
				},
				m.config.ExtraMounts...,
			),
			NetworkMode: "host",
			Args:        m.args(),
		},
	}

	initialClusterTokenArgument := "--initial-cluster-state=existing"
	if m.config.NewCluster {
		initialClusterTokenArgument = "--initial-cluster-token=etcd-cluster-2"
	}

	memberContainer.Config.Args = append(memberContainer.Config.Args, initialClusterTokenArgument)

	return &container.HostConfiguredContainer{
		Host:        m.config.Host,
		ConfigFiles: m.configFiles(),
		Container:   memberContainer,
	}, nil
}

func (m *member) peerAddress() string {
	return m.config.PeerAddress
}

// New validates MemberConfig and returns Member interface.
func (m *MemberConfig) New() (Member, error) {
	if err := m.Validate(); err != nil {
		return nil, fmt.Errorf("validating member configuration: %w", err)
	}

	nm := &member{
		config: m,
	}

	return nm, nil
}

// Validate validates etcd member configuration.
func (m *MemberConfig) Validate() error {
	var errors util.ValidateErrors

	nonEmptyFields := map[string]string{
		// TODO: Require peer address for now. Later we could figure out
		// how to use CNI for setting it using env variables or something.
		"peer address": m.PeerAddress,
		// TODO: Can we auto-generate it?
		"member name": m.Name,
	}

	for k, v := range nonEmptyFields {
		if v == "" {
			errors = append(errors, fmt.Errorf("%s can't be empty", k))
		}
	}

	certificates := map[string]string{
		"CA certificate":     m.CACertificate,
		"peer certificate":   m.PeerCertificate,
		"server certificate": m.ServerCertificate,
	}

	for certName, cert := range certificates {
		caCert := &pki.Certificate{
			X509Certificate: types.Certificate(cert),
		}

		if _, err := caCert.DecodeX509Certificate(); err != nil {
			errors = append(errors, fmt.Errorf("parsing %s as X.509 certificate: %w", certName, err))
		}
	}

	keys := map[string]string{
		"peer key":   m.PeerKey,
		"server key": m.ServerKey,
	}

	for k, v := range keys {
		if err := pki.ValidatePrivateKey(v); err != nil {
			errors = append(errors, fmt.Errorf("parsing %s as private key: %w", k, err))
		}
	}

	if err := m.Host.Validate(); err != nil {
		errors = append(errors, fmt.Errorf("validating host configuration: %w", err))
	}

	return errors.Return()
}

// peerURLs returns slice of peer urls assigned to member.
func (m *member) peerURLs() []string {
	return []string{fmt.Sprintf("https://%s", net.JoinHostPort(m.config.PeerAddress, "2380"))}
}

// forwardEndpoints opens forwarding connection for each endpoint
// and then returns new list of endpoints. If forwarding fails, error is returned.
func (m *member) forwardEndpoints(endpoints []string) ([]string, error) {
	newEndpoints := []string{}

	h, _ := m.config.Host.New() //nolint:errcheck // We check it in Validate().

	connectedHost, err := h.Connect()
	if err != nil {
		return nil, fmt.Errorf("opening forwarding connection to host: %w", err)
	}

	for _, e := range endpoints {
		e, err := connectedHost.ForwardTCP(e)
		if err != nil {
			return nil, fmt.Errorf("opening forwarding to member: %w", err)
		}

		newEndpoints = append(newEndpoints, fmt.Sprintf("https://%s", e))
	}

	return newEndpoints, nil
}

// getID returns etcd cluster member ID, based on either member name on the cluster or matching
// peer URL.
func (m *member) getID(cli etcdClient) (uint64, error) {
	// Get actual list of members.
	resp, err := cli.MemberList(context.Background())
	if err != nil {
		return 0, fmt.Errorf("listing existing cluster members: %w", err)
	}

	for _, member := range resp.Members {
		if member.Name == m.config.Name {
			return member.ID, nil
		}

		for _, p := range member.PeerURLs {
			for _, u := range m.peerURLs() {
				if p == u {
					return member.ID, nil
				}
			}
		}
	}

	return 0, nil
}

// getEtcdClient creates etcd client object using member certificates and
// given endpoints.
func (m *member) getEtcdClient(endpoints []string) (etcdClient, error) {
	//nolint:errcheck // We check it in Validate().
	cert, _ := tls.X509KeyPair([]byte(m.config.PeerCertificate), []byte(m.config.PeerKey))

	der, _ := pem.Decode([]byte(m.config.CACertificate))
	ca, _ := x509.ParseCertificate(der.Bytes) //nolint:errcheck // We check it in Validate().

	certPool := x509.NewCertPool()
	certPool.AddCert(ca)

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:            endpoints,
		DialTimeout:          defaultDialTimeout,
		DialKeepAliveTimeout: defaultDialTimeout,
		TLS: &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      certPool,
			MinVersion:   tls.VersionTLS12,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("creating etcd client: %w", err)
	}

	return cli, nil
}

// add uses given etcd client to add member into the cluster.
//
// If member is part of the cluster already, no error is returned.
func (m *member) add(cli etcdClient) error {
	memberID, err := m.getID(cli)
	if err != nil {
		return fmt.Errorf("getting member ID: %w", err)
	}

	// If no error is returned, and ID is 0, it means member is already added.
	if memberID != 0 {
		return nil
	}

	if _, err := cli.MemberAdd(context.Background(), m.peerURLs()); err != nil {
		return fmt.Errorf("adding new member to the cluster: %w", err)
	}

	return nil
}

// remove uses given etcd client to remove it from the cluster.
//
// If member is not part of the cluster anymore, no error is returned.
func (m *member) remove(cli etcdClient) error {
	memberID, err := m.getID(cli)
	if err != nil {
		return fmt.Errorf("getting member ID: %w", err)
	}

	// If no error is returned, and ID is 0, it means member is already returned.
	if memberID == 0 {
		return nil
	}

	if _, err = cli.MemberRemove(context.Background(), memberID); err != nil {
		return fmt.Errorf("removing member: %w", err)
	}

	return nil
}
