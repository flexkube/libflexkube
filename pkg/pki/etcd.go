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

	// User configurable fields, for easy use.

	CA        *Certificate      `json:"ca,omitempty"`
	Peers     map[string]string `json:"peers,omitempty"`
	Servers   map[string]string `json:"servers,omitempty"`
	ClientCNs []string          `json:"clientCNs,omitempty"`

	// Fields, where all certificates will be stored.

	PeerCertificates   map[string]*Certificate `json:"peerCertificates,omitempty"`
	ServerCertificates map[string]*Certificate `json:"serverCertificates,omitempty"`
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

	for k, v := range e.Peers {
		crs = append(crs, e.peerCR(k, v, defaultCertificate))
	}

	for k, v := range e.Servers {
		crs = append(crs, e.serverCR(k, v, defaultCertificate))
	}

	for _, k := range e.ClientCNs {
		crs = append(crs, e.clientCR(k, defaultCertificate))
	}

	return buildAndGenerate(crs...)
}

func (e *Etcd) peerCR(k, v string, defaultCertificate Certificate) *certificateRequest {
	if c := e.PeerCertificates[k]; c == nil {
		e.PeerCertificates[k] = &Certificate{}
	}

	return &certificateRequest{
		Target: e.PeerCertificates[k],
		CA:     e.CA,
		Certificates: []*Certificate{
			&defaultCertificate,
			&e.Certificate,
			clientServerCert(k, v),
			e.PeerCertificates[k],
		},
	}
}

func (e *Etcd) serverCR(k, v string, defaultCertificate Certificate) *certificateRequest {
	if c := e.ServerCertificates[k]; c == nil {
		e.ServerCertificates[k] = &Certificate{}
	}

	return &certificateRequest{
		Target: e.ServerCertificates[k],
		CA:     e.CA,
		Certificates: []*Certificate{
			&defaultCertificate,
			&e.Certificate,
			clientServerCert(k, v),
			e.ServerCertificates[k],
		},
	}
}

func clientServerCert(cn, ip string) *Certificate {
	return &Certificate{
		CommonName:  cn,
		IPAddresses: []string{ip, "127.0.0.1"},
		DNSNames:    []string{cn, "localhost"},
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
