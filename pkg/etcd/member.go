package etcd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"go.etcd.io/etcd/clientv3"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	containertypes "github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/types"
)

// Member represents single etcd member.
type Member struct {
	Name              string            `json:"name,omitempty"`
	Image             string            `json:"image,omitempty"`
	Host              host.Host         `json:"host,omitempty"`
	CACertificate     types.Certificate `json:"caCertificate,omitempty"`
	PeerCertificate   types.Certificate `json:"peerCertificate,omitempty"`
	PeerKey           types.PrivateKey  `json:"peerKey,omitempty"`
	PeerAddress       string            `json:"peerAddress,omitempty"`
	InitialCluster    string            `json:"initialCluster,omitempty"`
	PeerCertAllowedCN string            `json:"peerCertAllowedCN,omitempty"`
	ServerCertificate types.Certificate `json:"serverCertificate,omitempty"`
	ServerKey         types.PrivateKey  `json:"serverKey,omitempty"`
	ServerAddress     string            `json:"serverAddress,omitempty"`
	NewCluster        bool              `json:"newCluster,omitempty"`
}

// member is a validated, executable version of Member.
type member struct {
	name              string
	image             string
	host              host.Host
	caCertificate     string
	peerCertificate   string
	peerKey           string
	peerAddress       string
	initialCluster    string
	peerCertAllowedCN string
	serverCertificate string
	serverKey         string
	serverAddress     string
	newCluster        bool
}

func (m *member) configFiles() map[string]string {
	return map[string]string{
		"/etc/kubernetes/etcd/ca.crt":     m.caCertificate,
		"/etc/kubernetes/etcd/peer.crt":   m.peerCertificate,
		"/etc/kubernetes/etcd/peer.key":   m.peerKey,
		"/etc/kubernetes/etcd/server.crt": m.serverCertificate,
		"/etc/kubernetes/etcd/server.key": m.serverKey,
	}
}

// args returns flags which will be set to the container.
func (m *member) args() []string {
	return []string{
		// TODO Add descriptions explaining why we need each line.
		// Default value 'capnslog' for logger is deprecated and prints warning now.
		"--logger=zap", // Available only from 3.4.x
		// Since we are in container, listen on all interfaces.
		fmt.Sprintf("--listen-client-urls=https://%s:2379", m.serverAddress),
		fmt.Sprintf("--listen-peer-urls=https://%s:2380", m.peerAddress),
		fmt.Sprintf("--advertise-client-urls=https://%s:2379", m.serverAddress),
		fmt.Sprintf("--initial-advertise-peer-urls=https://%s:2380", m.peerAddress),
		fmt.Sprintf("--initial-cluster=%s", m.initialCluster),
		fmt.Sprintf("--name=%s", m.name),
		"--peer-trusted-ca-file=/etc/kubernetes/pki/etcd/ca.crt",
		"--peer-cert-file=/etc/kubernetes/pki/etcd/peer.crt",
		"--peer-key-file=/etc/kubernetes/pki/etcd/peer.key",
		"--peer-client-cert-auth",
		"--trusted-ca-file=/etc/kubernetes/pki/etcd/ca.crt",
		"--cert-file=/etc/kubernetes/pki/etcd/server.crt",
		"--key-file=/etc/kubernetes/pki/etcd/server.key",
		fmt.Sprintf("--data-dir=/%s.etcd", m.name),
		// To get rid of warning with default configuration.
		// ttl parameter support has been added in 3.4.x.
		"--auth-token=jwt,pub-key=/etc/kubernetes/pki/etcd/peer.crt,priv-key=/etc/kubernetes/pki/etcd/peer.key,sign-method=RS512,ttl=10m",
		// This is set by typhoon, seems like extra safety knob.
		"--strict-reconfig-check",
		// TODO: Enable metrics.
		// Enable TLS authentication with certificate CN field.
		// See https://github.com/etcd-io/etcd/blob/master/Documentation/op-guide/authentication.md#using-tls-common-name
		// for more details.
		"--client-cert-auth=true",
	}
}

// ToHostConfiguredContainer takes configured member and converts it to generic HostConfiguredContainer.
func (m *member) ToHostConfiguredContainer() (*container.HostConfiguredContainer, error) {
	c := container.Container{
		// TODO: This is weird. This sets docker as default runtime config.
		Runtime: container.RuntimeConfig{
			Docker: docker.DefaultConfig(),
		},
		Config: containertypes.ContainerConfig{
			Name:       fmt.Sprintf("etcd-%s", m.name),
			Image:      m.image,
			Entrypoint: []string{"/usr/local/bin/etcd"},
			Mounts: []containertypes.Mount{
				{
					// TODO: Between /var/lib/etcd and data dir we should probably put cluster name, to group them.
					// TODO: Make data dir configurable.
					Source: fmt.Sprintf("/var/lib/etcd/%s.etcd/", m.name),
					Target: fmt.Sprintf("/%s.etcd", m.name),
				},
				{
					Source: "/etc/kubernetes/etcd/",
					Target: "/etc/kubernetes/pki/etcd",
				},
			},
			NetworkMode: "host",
			Args:        m.args(),
		},
	}

	if m.newCluster {
		c.Config.Args = append(c.Config.Args, "--initial-cluster-token=etcd-cluster-2")
	} else {
		c.Config.Args = append(c.Config.Args, "--initial-cluster-state=existing")
	}

	return &container.HostConfiguredContainer{
		Host:        m.host,
		ConfigFiles: m.configFiles(),
		Container:   c,
	}, nil
}

// New validates Member configuration and returns it's usable version.
func (m *Member) New() (container.ResourceInstance, error) {
	if err := m.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate member configuration: %w", err)
	}

	nm := &member{
		name:              m.Name,
		image:             m.Image,
		host:              m.Host,
		caCertificate:     string(m.CACertificate),
		peerCertificate:   string(m.PeerCertificate),
		peerKey:           string(m.PeerKey),
		peerAddress:       m.PeerAddress,
		initialCluster:    m.InitialCluster,
		peerCertAllowedCN: m.PeerCertAllowedCN,
		serverCertificate: string(m.ServerCertificate),
		serverKey:         string(m.ServerKey),
		serverAddress:     m.ServerAddress,
		newCluster:        m.NewCluster,
	}

	return nm, nil
}

// Validate validates etcd member configuration.
//
// TODO: Add validation of certificates if specified.
func (m *Member) Validate() error {
	var errors util.ValidateError

	// TODO: Require peer address for now. Later we could figure out
	// how to use CNI for setting it using env variables or something.
	if m.PeerAddress == "" {
		errors = append(errors, fmt.Errorf("peer address must be set"))
	}

	// TODO: Can we auto-generate it?
	if m.Name == "" {
		errors = append(errors, fmt.Errorf("member name must be set"))
	}

	if err := m.Host.Validate(); err != nil {
		errors = append(errors, fmt.Errorf("host validation failed: %w", err))
	}

	return errors.Return()
}

func (m *member) peerURLs() []string {
	return []string{fmt.Sprintf("https://%s:2380", m.peerAddress)}
}

// forwardEndpoints opens forwarding connection for each endpoint
// and then returns new list of endpoints. If forwarding fails, error is returned.
func (m *member) forwardEndpoints(endpoints []string) ([]string, error) {
	newEndpoints := []string{}

	h, _ := m.host.New()

	hc, err := h.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed opening forwarding connection to host: %w", err)
	}

	for _, e := range endpoints {
		e, err := hc.ForwardTCP(e)
		if err != nil {
			return nil, fmt.Errorf("failed opening forwarding to member: %w", err)
		}

		newEndpoints = append(newEndpoints, fmt.Sprintf("https://%s", e))
	}

	return newEndpoints, nil
}

func (m *member) getID(cli etcdClient) (uint64, error) {
	// Get actual list of members.
	resp, err := cli.MemberList(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to list existing cluster members: %w", err)
	}

	for _, v := range resp.Members {
		if v.Name == m.name {
			return v.ID, nil
		}

		for _, p := range v.PeerURLs {
			for _, u := range m.peerURLs() {
				if p == u {
					return v.ID, nil
				}
			}
		}
	}

	return 0, nil
}

func (m *member) getEtcdClient(endpoints []string) (etcdClient, error) {
	cert, _ := tls.X509KeyPair([]byte(m.peerCertificate), []byte(m.peerKey))
	der, _ := pem.Decode([]byte(m.caCertificate))
	ca, _ := x509.ParseCertificate(der.Bytes)

	p := x509.NewCertPool()
	p.AddCert(ca)

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:            endpoints,
		DialTimeout:          defaultDialTimeout,
		DialKeepAliveTimeout: defaultDialTimeout,
		TLS: &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      p,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating etcd client: %w", err)
	}

	return cli, nil
}

func (m *member) add(cli etcdClient) error {
	id, err := m.getID(cli)
	if err != nil {
		return fmt.Errorf("failed getting member ID: %w", err)
	}

	// If no error is returned, and ID is 0, it means member is already returned.
	if id != 0 {
		return nil
	}

	if _, err := cli.MemberAdd(context.Background(), m.peerURLs()); err != nil {
		return fmt.Errorf("failed adding new member to the cluster: %w", err)
	}

	return nil
}

func (m *member) remove(cli etcdClient) error {
	id, err := m.getID(cli)
	if err != nil {
		return fmt.Errorf("failed getting member ID: %w", err)
	}

	// If no error is returned, and ID is 0, it means member is already returned.
	if id == 0 {
		return nil
	}

	if _, err = cli.MemberRemove(context.Background(), id); err != nil {
		return fmt.Errorf("failed removing member: %w", err)
	}

	return nil
}
