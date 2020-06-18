package pki

import (
	"fmt"
)

const (
	// EtcdCACN is a default CN for etcd CA certificate, as recommended by
	// https://kubernetes.io/docs/setup/best-practices/certificates/.
	EtcdCACN = "etcd-ca"
)

// Etcd stores etcd PKI and their settings.
type Etcd struct {
	// Inline Certificate struct, so some settings can be applied as defaults for all etcd certificates.
	Certificate

	// CA stores etcd CA certificate.
	CA *Certificate `json:"ca,omitempty"`

	// Peers is a map of peer certificates to generate, where key is name of the peer and value
	// is the IP address on which peer will be listening on.
	Peers map[string]string `json:"peers,omitempty"`

	// Servers is a map of server certificates to generate, where key is the CN of the client
	// certificate and value is the IP address on which the server will be listening on.
	Servers map[string]string `json:"servers,omitempty"`

	// ClientCNS is a list of client certificate Common Names to generate.
	ClientCNs []string `json:"clientCNs,omitempty"`

	// PeerCertificates defines and stores all peer certificates.
	PeerCertificates map[string]*Certificate `json:"peerCertificates,omitempty"`

	// ServerCertificates defines and stores all server certificates.
	ServerCertificates map[string]*Certificate `json:"serverCertificates,omitempty"`

	// ClientCertificates defined and stores all client certificates.
	ClientCertificates map[string]*Certificate `json:"clientCertificates,omitempty"`
}

// Generate generates etcd PKI.
func (e *Etcd) Generate(rootCA *Certificate, defaultCertificate Certificate) error {
	if e.CA == nil {
		e.CA = &Certificate{}
	}

	if e.Peers != nil && e.PeerCertificates == nil {
		e.PeerCertificates = map[string]*Certificate{}
	}

	// If there is no different server certificates defined, assume they are the same as peers.
	if e.Servers == nil && e.Peers != nil {
		e.Servers = e.Peers
	}

	if e.Servers != nil && e.ServerCertificates == nil {
		e.ServerCertificates = map[string]*Certificate{}
	}

	if len(e.ClientCNs) != 0 && e.ClientCertificates == nil {
		e.ClientCertificates = map[string]*Certificate{}
	}

	cr := &certificateRequest{
		Target: e.CA,
		CA:     rootCA,
		Certificates: []*Certificate{
			&defaultCertificate,
			&e.Certificate,
			caCertificate(EtcdCACN),
			e.CA,
		},
	}

	// etcd CA Certificate
	if err := buildAndGenerate(cr); err != nil {
		return fmt.Errorf("failed to generate etcd CA certificate: %w", err)
	}

	crs := []*certificateRequest{}

	for commonName, ip := range e.Peers {
		crs = append(crs, e.peerCR(commonName, ip, defaultCertificate))
	}

	for commonName, ip := range e.Servers {
		crs = append(crs, e.serverCR(commonName, ip, defaultCertificate))
	}

	for _, commonName := range e.ClientCNs {
		crs = append(crs, e.clientCR(commonName, defaultCertificate))
	}

	return buildAndGenerate(crs...)
}

func (e *Etcd) peerCR(commonName, ip string, defaultCertificate Certificate) *certificateRequest {
	if c := e.PeerCertificates[commonName]; c == nil {
		e.PeerCertificates[commonName] = &Certificate{}
	}

	return &certificateRequest{
		Target: e.PeerCertificates[commonName],
		CA:     e.CA,
		Certificates: []*Certificate{
			&defaultCertificate,
			&e.Certificate,
			clientServerCert(commonName, ip),
			e.PeerCertificates[commonName],
		},
	}
}

func (e *Etcd) serverCR(commonName, ip string, defaultCertificate Certificate) *certificateRequest {
	if c := e.ServerCertificates[commonName]; c == nil {
		e.ServerCertificates[commonName] = &Certificate{}
	}

	return &certificateRequest{
		Target: e.ServerCertificates[commonName],
		CA:     e.CA,
		Certificates: []*Certificate{
			&defaultCertificate,
			&e.Certificate,
			clientServerCert(commonName, ip),
			e.ServerCertificates[commonName],
		},
	}
}

func clientServerCert(commonName, ip string) *Certificate {
	return &Certificate{
		CommonName:  commonName,
		IPAddresses: []string{ip, "127.0.0.1"},
		DNSNames:    []string{commonName, "localhost"},
		KeyUsage:    clientServerUsage(),
	}
}

func (e *Etcd) clientCR(k string, defaultCertificate Certificate) *certificateRequest {
	if c := e.ClientCertificates[k]; c == nil {
		e.ClientCertificates[k] = &Certificate{}
	}

	return &certificateRequest{
		Target: e.ClientCertificates[k],
		CA:     e.CA,
		Certificates: []*Certificate{
			&defaultCertificate,
			&e.Certificate,
			{
				CommonName: k,
				KeyUsage:   clientUsage(),
			},
			e.ClientCertificates[k],
		},
	}
}
