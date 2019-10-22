package etcd

import (
	"fmt"

	"github.com/invidian/flexkube/pkg/container"
	"github.com/invidian/flexkube/pkg/container/runtime/docker"
	"github.com/invidian/flexkube/pkg/container/types"
	"github.com/invidian/flexkube/pkg/defaults"
	"github.com/invidian/flexkube/pkg/host"
)

// Member represents single etcd member
type Member struct {
	Name              string     `json:"name,omitempty" yaml:"name,omitempty"`
	Image             string     `json:"image,omitempty" yaml:"image,omitempty"`
	Host              *host.Host `json:"host,omitempty" yaml:"host,omitempty"`
	PeerCACertificate string     `json:"peerCACertificate,omitempty" yaml:"peerCACertificate,omitempty"`
	PeerCertificate   string     `json:"peerCertificate,omitempty" yaml:"peerCertificate,omitempty"`
	PeerKey           string     `json:"peerKey,omitempty" yaml:"peerKey,omitempty"`
	PeerAddress       string     `json:"peerAddress,omitempty" yaml:"peerAddress,omitempty"`
	InitialCluster    string     `json:"initialCluster,omitempty" yaml:"initialCluster,omitempty"`
	PeerCertAllowedCN string     `json:"peerCertAllowedCN,omitempty" yaml:"peerCertAllowedCN,omitempty"`
}

// member is a validated, executable version of Member
type member struct {
	name              string
	image             string
	host              host.Host
	peerCACertificate string
	peerCertificate   string
	peerKey           string
	peerAddress       string
	initialCluster    string
	peerCertAllowedCN string
}

func (m *member) ToHostConfiguredContainer() *container.HostConfiguredContainer {
	configFiles := make(map[string]string)
	if m.peerCACertificate != "" && m.peerCertificate != "" && m.peerCertificate != "" {
		configFiles["/etc/kubernetes/pki/etcd/ca.crt"] = m.peerCACertificate
		configFiles["/etc/kubernetes/pki/etcd/peer.crt"] = m.peerCertificate
		configFiles["/etc/kubernetes/pki/etcd/peer.key"] = m.peerKey
	}

	c := container.Container{
		// TODO this is weird. This sets docker as default runtime config
		Runtime: container.RuntimeConfig{
			Docker: &docker.ClientConfig{},
		},
		Config: types.ContainerConfig{
			Name:       m.name,
			Image:      m.image,
			Entrypoint: []string{"/usr/local/bin/etcd"},
			Ports: []types.PortMap{
				types.PortMap{
					IP:       m.peerAddress,
					Protocol: "tcp",
					Port:     2379,
				},
				types.PortMap{
					IP:       m.peerAddress,
					Protocol: "tcp",
					Port:     2380,
				},
			},
			Mounts: []types.Mount{
				types.Mount{
					// TODO between /var/lib/etcd and data dir we should probably put cluster name, to group them
					// TODO make data dir configurable
					Source: fmt.Sprintf("/var/lib/etcd/%s.etcd/", m.name),
					Target: fmt.Sprintf("/%s.etcd", m.name),
				},
				types.Mount{
					Source: "/etc/kubernetes/pki/etcd/",
					Target: "/etc/kubernetes/pki/etcd",
				},
			},
			Args: []string{
				//TODO Add descriptions explaining why we need each line.
				// Default value 'capnslog' for logger is deprecated and prints warning now.
				"--logger=zap",
				// Since we are in container, listen on all interfaces
				"--listen-client-urls=http://0.0.0.0:2379",
				"--listen-peer-urls=https://0.0.0.0:2380",
				fmt.Sprintf("--advertise-client-urls=http://%s:2379", m.peerAddress),
				fmt.Sprintf("--initial-advertise-peer-urls=https://%s:2380", m.peerAddress),
				fmt.Sprintf("--initial-cluster=%s", m.initialCluster),
				fmt.Sprintf("--name=%s", m.name),
				"--initial-cluster-token=etcd-cluster-2",
				"--peer-trusted-ca-file=/etc/kubernetes/pki/etcd/ca.crt",
				"--peer-cert-file=/etc/kubernetes/pki/etcd/peer.crt",
				"--peer-key-file=/etc/kubernetes/pki/etcd/peer.key",
				"--peer-client-cert-auth",
				fmt.Sprintf("--data-dir=/%s.etcd", m.name),
				// To get rid of warning with default configuration
				"--auth-token=jwt,pub-key=/etc/kubernetes/pki/etcd/peer.crt,priv-key=/etc/kubernetes/pki/etcd/peer.key,sign-method=RS512,ttl=10m",
				// This is set by typhoon, seems like extra safety knob
				"--strict-reconfig-check",
				// TODO enable metrics
			},
		},
	}

	return &container.HostConfiguredContainer{
		Host:        m.host,
		ConfigFiles: configFiles,
		Container:   c,
	}
}

func (m *member) ToExistingClusterMember() *container.HostConfiguredContainer {
	return nil
}

func (m *member) ToNewClusterMember() *container.HostConfiguredContainer {
	return nil
}

func (m *Member) New() (*member, error) {
	if err := m.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate member configuration: %w", err)
	}

	nm := &member{
		name:              m.Name,
		image:             m.Image,
		host:              *m.Host,
		peerCACertificate: m.PeerCACertificate,
		peerCertificate:   m.PeerCertificate,
		peerKey:           m.PeerKey,
		peerAddress:       m.PeerAddress,
		initialCluster:    m.InitialCluster,
		peerCertAllowedCN: m.PeerCertAllowedCN,
	}

	if nm.image == "" {
		nm.image = defaults.EtcdImage
	}

	return nm, nil
}

// TODO add validation of certificates if specified
func (m *Member) Validate() error {
	// TODO require peer address for now. Later we could figure out
	// how to use CNI for setting it using env variables or something
	if m.PeerAddress == "" {
		return fmt.Errorf("peer address must be set")
	}

	// TODO can we auto-generate it?
	if m.Name == "" {
		return fmt.Errorf("member name must be set")
	}

	// TODO actually direct, local container is fine too, this check can be removed
	if m.Host == nil {
		return fmt.Errorf("host configuration must be defined")
	}

	return nil
}
