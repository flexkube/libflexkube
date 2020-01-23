package etcd

import (
	"fmt"

	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	containertypes "github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/defaults"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/types"
)

// Member represents single etcd member
type Member struct {
	Name              string            `json:"name"`
	Image             string            `json:"image"`
	Host              host.Host         `json:"host"`
	CACertificate     types.Certificate `json:"caCertificate"`
	PeerCertificate   types.Certificate `json:"peerCertificate"`
	PeerKey           types.PrivateKey  `json:"peerKey"`
	PeerAddress       string            `json:"peerAddress"`
	InitialCluster    string            `json:"initialCluster"`
	PeerCertAllowedCN string            `json:"peerCertAllowedCN"`
	ServerCertificate types.Certificate `json:"serverCertificate"`
	ServerKey         types.PrivateKey  `json:"serverKey"`
	ServerAddress     string            `json:"serverAddress"`
}

// member is a validated, executable version of Member
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

// ToHostConfiguredContainer takes configured member and converts it to generic HostConfiguredContainer
func (m *member) ToHostConfiguredContainer() (*container.HostConfiguredContainer, error) {
	c := container.Container{
		// TODO this is weird. This sets docker as default runtime config
		Runtime: container.RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: containertypes.ContainerConfig{
			Name:       m.name,
			Image:      m.image,
			Entrypoint: []string{"/usr/local/bin/etcd"},
			Mounts: []containertypes.Mount{
				{
					// TODO between /var/lib/etcd and data dir we should probably put cluster name, to group them
					// TODO make data dir configurable
					Source: fmt.Sprintf("/var/lib/etcd/%s.etcd/", m.name),
					Target: fmt.Sprintf("/%s.etcd", m.name),
				},
				{
					Source: "/etc/kubernetes/etcd/",
					Target: "/etc/kubernetes/pki/etcd",
				},
			},
			NetworkMode: "host",
			Args: []string{
				//TODO Add descriptions explaining why we need each line.
				// Default value 'capnslog' for logger is deprecated and prints warning now.
				//"--logger=zap", // Available only from 3.4.x
				// Since we are in container, listen on all interfaces
				fmt.Sprintf("--listen-client-urls=https://%s:2379", m.serverAddress),
				fmt.Sprintf("--listen-peer-urls=https://%s:2380", m.peerAddress),
				fmt.Sprintf("--advertise-client-urls=https://%s:2379", m.serverAddress),
				fmt.Sprintf("--initial-advertise-peer-urls=https://%s:2380", m.peerAddress),
				fmt.Sprintf("--initial-cluster=%s", m.initialCluster),
				fmt.Sprintf("--name=%s", m.name),
				"--initial-cluster-token=etcd-cluster-2",
				"--peer-trusted-ca-file=/etc/kubernetes/pki/etcd/ca.crt",
				"--peer-cert-file=/etc/kubernetes/pki/etcd/peer.crt",
				"--peer-key-file=/etc/kubernetes/pki/etcd/peer.key",
				"--peer-client-cert-auth",
				"--trusted-ca-file=/etc/kubernetes/pki/etcd/ca.crt",
				"--cert-file=/etc/kubernetes/pki/etcd/server.crt",
				"--key-file=/etc/kubernetes/pki/etcd/server.key",
				fmt.Sprintf("--data-dir=/%s.etcd", m.name),
				// To get rid of warning with default configuration.
				// ttl parameter support has been added in 3.4.x
				"--auth-token=jwt,pub-key=/etc/kubernetes/pki/etcd/peer.crt,priv-key=/etc/kubernetes/pki/etcd/peer.key,sign-method=RS512,ttl=10m",
				// This is set by typhoon, seems like extra safety knob
				"--strict-reconfig-check",
				// TODO enable metrics
			},
		},
	}

	return &container.HostConfiguredContainer{
		Host:        m.host,
		ConfigFiles: m.configFiles(),
		Container:   c,
	}, nil
}

// New valides Member configuration and returns it's usable version
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
	}

	if nm.image == "" {
		nm.image = defaults.EtcdImage
	}

	return nm, nil
}

// Validate validates etcd member configuration
// TODO add validation of certificates if specified
func (m *Member) Validate() error {
	var errors types.ValidateError

	// TODO require peer address for now. Later we could figure out
	// how to use CNI for setting it using env variables or something
	if m.PeerAddress == "" {
		errors = append(errors, fmt.Errorf("peer address must be set"))
	}

	// TODO can we auto-generate it?
	if m.Name == "" {
		errors = append(errors, fmt.Errorf("member name must be set"))
	}

	if err := m.Host.Validate(); err != nil {
		errors = append(errors, fmt.Errorf("host validation failed: %w", err))
	}

	return errors.Return()
}
