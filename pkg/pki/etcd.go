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

	servers := e.Servers

	// If there is no different server certificates defined, assume they are the same as peers.
	if e.Servers == nil && e.Peers != nil {
		servers = e.Peers
	}

	if e.PeerCertificates == nil && len(e.Peers) != 0 {
		e.PeerCertificates = map[string]*Certificate{}
	}

	if e.ServerCertificates == nil && len(servers) != 0 {
		e.ServerCertificates = map[string]*Certificate{}
	}

	if e.ClientCertificates == nil && len(e.ClientCNs) != 0 {
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

	crs = append(crs, e.crsFromMap(&defaultCertificate, e.PeerCertificates, e.Peers, true)...)
	crs = append(crs, e.crsFromMap(&defaultCertificate, e.ServerCertificates, servers, true)...)

	clientCNsMap := map[string]string{}
	for _, commonName := range e.ClientCNs {
		clientCNsMap[commonName] = ""
	}

	crs = append(crs, e.crsFromMap(&defaultCertificate, e.ClientCertificates, clientCNsMap, false)...)

	return buildAndGenerate(crs...)
}

// certificateFromCNIPMap produces a certificate from given common name and IP address.
func certificateFromCNIPMap(commonName string, ip string, server bool) *Certificate {
	c := &Certificate{
		CommonName: commonName,
		KeyUsage:   clientUsage(),
	}

	if server {
		c.KeyUsage = clientServerUsage()
		c.DNSNames = []string{commonName, "localhost"}
	}

	if ip != "" && server {
		c.IPAddresses = append(c.IPAddresses, ip, "127.0.0.1")
	}

	return c
}

// peerCRs builds list of certificate requests for peer certificates by combining
// information from PeerCertificates and Peers fields, where PeerCertificates always
// takes precedence.
func (e *Etcd) crsFromMap(defaultCertificate *Certificate, certs map[string]*Certificate, cnIPs map[string]string, server bool) []*certificateRequest {
	// Store peer CRs in temporary map, so we can find them by common name.
	crs := map[string]*certificateRequest{}

	// Iterate over peer certificates, as they should take priority over
	// Peers field.
	for commonName := range certs {
		crs[commonName] = &certificateRequest{
			Target: certs[commonName],
			CA:     e.CA,
			Certificates: []*Certificate{
				defaultCertificate,
				&e.Certificate,
				certificateFromCNIPMap(commonName, cnIPs[commonName], server),
				certs[commonName],
			},
		}
	}

	for commonName, ip := range cnIPs {
		// If certificate request is already created for a given common name, it will
		// have peers information included, so we jump to another one.
		if _, ok := crs[commonName]; ok {
			continue
		}

		// Make sure target certificate is initialized.
		if _, ok := certs[commonName]; !ok {
			certs[commonName] = &Certificate{}
		}

		crs[commonName] = &certificateRequest{
			Target: certs[commonName],
			CA:     e.CA,
			Certificates: []*Certificate{
				defaultCertificate,
				&e.Certificate,
				certificateFromCNIPMap(commonName, ip, server),
			},
		}
	}

	return certificateRequestsFromMap(crs)
}

func certificateRequestsFromMap(crsMap map[string]*certificateRequest) []*certificateRequest {
	crs := []*certificateRequest{}

	for _, cr := range crsMap {
		crs = append(crs, cr)
	}

	return crs
}
