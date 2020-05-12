// Package pki allows to manage Kubernetes PKI certificates.
package pki

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/pkg/types"
)

const (
	// RSABits is a default private key length. Default is 2048, as it's quite secure and generating
	// 4096 keys takes a lot of time and increases generation time by the factor of 10. Once generation
	// process is done in parallel, it should be increased.
	RSABits = 2048

	// Organization is a default organization name in generated certificates.
	Organization = "organization"

	// ValidityDuration is a default time the certificates are valid. Defaults to 365 days.
	ValidityDuration = "8760h"

	// RenewThreshold defines minimum remaining validity time for the certificate, before
	// is will be renewed.
	RenewThreshold = "720h"

	// X509CertificatePEMHeader is a PEM format header used while encoding X.509 certificates.
	X509CertificatePEMHeader = "CERTIFICATE"

	// RSAPrivateKeyPEMHeader is a PEM format header user while encoding RSA private keys.
	RSAPrivateKeyPEMHeader = "RSA PRIVATE KEY"

	// RSAPublicKeyPEMHeader is a PEM format header user while encoding RSA public keys.
	RSAPublicKeyPEMHeader = "RSA PUBLIC KEY"

	// RootCACN is a default CN for root CA certificate.
	RootCACN = "root-ca"
)

func keyUsage(k string) x509.KeyUsage {
	return map[string]x509.KeyUsage{
		"digital_signature":  x509.KeyUsageDigitalSignature,
		"content_commitment": x509.KeyUsageContentCommitment,
		"key_encipherment":   x509.KeyUsageKeyEncipherment,
		"data_encipherment":  x509.KeyUsageDataEncipherment,
		"key_agreement":      x509.KeyUsageKeyAgreement,
		"cert_signing":       x509.KeyUsageCertSign,
		"crl_signing":        x509.KeyUsageCRLSign,
		"encipher_only":      x509.KeyUsageEncipherOnly,
		"decipher_only":      x509.KeyUsageDecipherOnly,
	}[k]
}

func extKeyUsage(k string) x509.ExtKeyUsage {
	return map[string]x509.ExtKeyUsage{
		"any_extended":                  x509.ExtKeyUsageAny,
		"server_auth":                   x509.ExtKeyUsageServerAuth,
		"client_auth":                   x509.ExtKeyUsageClientAuth,
		"code_signing":                  x509.ExtKeyUsageCodeSigning,
		"email_protection":              x509.ExtKeyUsageEmailProtection,
		"ipsec_end_system":              x509.ExtKeyUsageIPSECEndSystem,
		"ipsec_tunnel":                  x509.ExtKeyUsageIPSECTunnel,
		"ipsec_user":                    x509.ExtKeyUsageIPSECUser,
		"timestamping":                  x509.ExtKeyUsageTimeStamping,
		"ocsp_signing":                  x509.ExtKeyUsageOCSPSigning,
		"microsoft_server_gated_crypto": x509.ExtKeyUsageMicrosoftServerGatedCrypto,
		"netscape_server_gated_crypto":  x509.ExtKeyUsageNetscapeServerGatedCrypto,
	}[k]
}

// Certificate defines configurable options for each certificate.
type Certificate struct {
	Organization     string            `json:"organization,omitempty"`
	RSABits          int               `json:"rsaBits,omitempty"`
	ValidityDuration string            `json:"validityDuration,omitempty"`
	RenewThreshold   string            `json:"renewThreshold,omitempty"`
	CommonName       string            `json:"commonName,omitempty"`
	CA               bool              `json:"ca,omitempty"`
	KeyUsage         []string          `json:"keyUsage,omitempty"`
	IPAddresses      []string          `json:"ipAddresses,omitempty"`
	DNSNames         []string          `json:"dnsNames,omitempty"`
	X509Certificate  types.Certificate `json:"x509Certificate,omitempty"`
	PublicKey        string            `json:"publicKey,omitempty"`
	PrivateKey       types.PrivateKey  `json:"privateKey,omitempty"`
}

// PKI contains configuration and all generated certificates and private keys required for running Kubernetes.
type PKI struct {
	// Inline Certificate struct, so some settings can be applied as defaults for all certificates in PKI.
	Certificate

	// RootCA contains configuration and generated root CA certificate and private key.
	RootCA *Certificate `json:"rootCA,omitempty"`

	// Etcd contains configuration and generated all etcd certificates and private keys.
	Etcd *Etcd `json:"etcd,omitempty"`

	// Kubernetes contains configuration and generated all Kubernetes certificates and private keys.
	Kubernetes *Kubernetes `json:"kubernetes,omitempty"`
}

func serverUsage() []string {
	return []string{
		"key_encipherment",
		"digital_signature",
		"server_auth",
	}
}

func clientUsage() []string {
	return []string{
		"key_encipherment",
		"digital_signature",
		"client_auth",
	}
}

func clientServerUsage() []string {
	return []string{
		"key_encipherment",
		"digital_signature",
		"client_auth",
		"server_auth",
	}
}

func caUsage() []string {
	return []string{
		"key_encipherment",
		"digital_signature",
		"cert_signing",
	}
}

func caCertificate(cn string) *Certificate {
	return &Certificate{
		CommonName: cn,
		CA:         true,
		KeyUsage:   caUsage(),
	}
}

type certificateRequest struct {
	Target       *Certificate
	CA           *Certificate
	Certificates []*Certificate
}

func buildAndGenerate(crs ...*certificateRequest) error {
	for _, cr := range crs {
		r, err := buildCertificate(cr.Certificates...)
		if err != nil {
			return fmt.Errorf("failed to build certificate configuration: %w", err)
		}

		if err := r.Generate(cr.CA); err != nil {
			return fmt.Errorf("failed to generate the certificate: %w", err)
		}

		*cr.Target = *r
	}

	return nil
}

func (p *PKI) generateRootCA() error {
	if p.RootCA == nil {
		p.RootCA = &Certificate{}
	}

	cr := &certificateRequest{
		Target: p.RootCA,
		Certificates: []*Certificate{
			&p.Certificate,
			caCertificate(RootCACN),
			p.RootCA,
		},
	}

	if err := buildAndGenerate(cr); err != nil {
		return fmt.Errorf("failed to generate root CA certificate: %w", err)
	}

	return nil
}

// Generate generates PKI required for running Kubernetes, including root CA and etcd certificates.
func (p *PKI) Generate() error {
	if err := p.generateRootCA(); err != nil {
		return fmt.Errorf("failed to generate root CA certificate: %w", err)
	}

	// If etcd field is set, generate etcd PKI. This allows to skip generation of those certificates,
	// if one deploys just Kubernetes on existing etcd cluster.
	if p.Etcd != nil {
		if err := p.Etcd.Generate(p.RootCA, p.Certificate); err != nil {
			return fmt.Errorf("failed to generate etcd PKI: %w", err)
		}
	}

	// If Kubernetes field is set, generate Kubernetes PKI. This allows to skip generation of those certificates,
	// if one deploys just etcd cluster.
	if p.Kubernetes != nil {
		if err := p.Kubernetes.Generate(p.RootCA, p.Certificate); err != nil {
			return fmt.Errorf("failed to generate Kubernetes PKI: %w", err)
		}
	}

	return nil
}

// buildCertificate merges N number of given certificates. Properties of last given certificate takes
// precedence over previous ones.
func buildCertificate(certs ...*Certificate) (*Certificate, error) {
	r := &Certificate{
		Organization:     Organization,
		RSABits:          RSABits,
		ValidityDuration: ValidityDuration,
		RenewThreshold:   RenewThreshold,
	}

	for _, c := range certs {
		rc, err := yaml.Marshal(c)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal certificate: %w", err)
		}

		if err := yaml.Unmarshal(rc, r); err != nil {
			return nil, fmt.Errorf("failed to unmarshal the certificate: %w", err)
		}
	}

	return r, nil
}

func (c *Certificate) decodePrivateKey() (*rsa.PrivateKey, error) {
	der, _ := pem.Decode([]byte(c.PrivateKey))
	if der == nil {
		return nil, fmt.Errorf("private key is not defined in valid PEM format")
	}

	k, err := x509.ParsePKCS1PrivateKey(der.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key to PKCS1 format: %w", err)
	}

	return k, nil
}

func (c *Certificate) decodeX509Certificate() (*x509.Certificate, error) {
	der, _ := pem.Decode([]byte(c.X509Certificate))
	if der == nil {
		return nil, fmt.Errorf("X.509 certificate is not defined in valid PEM format") //nolint:stylecheck
	}

	cert, err := x509.ParseCertificate(der.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse X.509 certificate: %w", err)
	}

	return cert, nil
}

// persistPublicKey persist given RSA public key into the certificate object.
func (c *Certificate) persistPublicKey(k *rsa.PublicKey) error {
	pubBytes, err := x509.MarshalPKIXPublicKey(k)
	if err != nil {
		return fmt.Errorf("failed marshaling RSA public key: %w", err)
	}

	var buf bytes.Buffer

	if err := pem.Encode(&buf, &pem.Block{Type: RSAPublicKeyPEMHeader, Bytes: pubBytes}); err != nil {
		return fmt.Errorf("failed to encode RSA public key: %w", err)
	}

	c.PublicKey = buf.String()

	return nil
}

func (c *Certificate) generatePrivateKey() (*rsa.PrivateKey, error) {
	// generate RSA private key.
	k, err := rsa.GenerateKey(rand.Reader, c.RSABits)
	if err != nil {
		return nil, fmt.Errorf("failed generating RSA key: %w", err)
	}

	privBytes := x509.MarshalPKCS1PrivateKey(k)

	var buf bytes.Buffer
	if err := pem.Encode(&buf, &pem.Block{Type: RSAPrivateKeyPEMHeader, Bytes: privBytes}); err != nil {
		return nil, fmt.Errorf("failed to encode RSA private key: %w", err)
	}

	c.PrivateKey = types.PrivateKey(buf.String())

	if err := c.persistPublicKey(k.Public().(*rsa.PublicKey)); err != nil {
		return nil, fmt.Errorf("failed persisting RSA public key: %w", err)
	}

	return k, nil
}

func (c *Certificate) getPrivateKey() (*rsa.PrivateKey, error) {
	if c.PrivateKey != "" {
		return c.decodePrivateKey()
	}

	return c.generatePrivateKey()
}

// Validate validates the certificate configuration.
func (c *Certificate) Validate() error {
	if _, err := time.ParseDuration(c.ValidityDuration); err != nil {
		return fmt.Errorf("failed to parse validity duration %q for certificate: %w", c.ValidityDuration, err)
	}

	for _, i := range c.IPAddresses {
		if ip := net.ParseIP(i); ip == nil {
			return fmt.Errorf("failed parsing IP address %q", i)
		}
	}

	if c.RSABits == 0 {
		return fmt.Errorf("RSA bits can't be 0")
	}

	return nil
}

func (c *Certificate) decodeKeyUsage() (x509.KeyUsage, []x509.ExtKeyUsage) {
	ku := 0
	eku := []x509.ExtKeyUsage{}

	for _, k := range c.KeyUsage {
		r := int(keyUsage(k))
		if r != 0 {
			ku |= r
			continue
		}

		if e := extKeyUsage(k); e != 0 {
			eku = append(eku, e)
		}
	}

	return x509.KeyUsage(ku), eku
}

func (c *Certificate) generateX509Certificate(k *rsa.PrivateKey, ca *Certificate) error {
	// Generate serial number for X.509 certificate.
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	vd, _ := time.ParseDuration(c.ValidityDuration)

	ku, eku := c.decodeKeyUsage()

	cert := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{c.Organization},
			CommonName:   c.CommonName,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(vd),

		KeyUsage:              ku,
		ExtKeyUsage:           eku,
		BasicConstraintsValid: true,
		IsCA:                  c.CA,
		DNSNames:              c.DNSNames,
	}

	for _, i := range c.IPAddresses {
		cert.IPAddresses = append(cert.IPAddresses, net.ParseIP(i))
	}

	pk := k
	caCert := &cert

	if ca != nil {
		var err error

		caCert, pk, err = ca.decodeKeypair()
		if err != nil {
			return fmt.Errorf("failed to decode CA keypair: %w", err)
		}
	}

	der, err := x509.CreateCertificate(rand.Reader, &cert, caCert, &k.PublicKey, pk)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	return c.persistX509Certificate(der)
}

// decodeKeypair decodes both X.509 certificate and private key.
func (c *Certificate) decodeKeypair() (*x509.Certificate, *rsa.PrivateKey, error) {
	pk, err := c.decodePrivateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	cert, err := c.decodeX509Certificate()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode X.509 certificate: %w", err)
	}

	return cert, pk, nil
}

func (c *Certificate) persistX509Certificate(der []byte) error {
	// Encode private certificate from PEM into PEM format.
	var cert bytes.Buffer

	if err := pem.Encode(&cert, &pem.Block{Type: X509CertificatePEMHeader, Bytes: der}); err != nil {
		return fmt.Errorf("failed to write data to cert.pem: %w", err)
	}

	c.X509Certificate = types.Certificate(cert.String())

	return nil
}

// Generate ensures that all fields of the certificate are populated.
//
// This function currently supports:
// - Generating new RSA private key and public key.
// - Generating new X.509 certificates.
//
// NOT implemented functionality:
// - Renewing certificates based on expiry time.
// - Renewing X.509 certificate after RSA private key renewal.
// - Renewing issued certificate during CA renewal.
func (c *Certificate) Generate(ca *Certificate) error {
	if err := c.Validate(); err != nil {
		return fmt.Errorf("failed validating the certificate: %w", err)
	}

	k, err := c.getPrivateKey()
	if err != nil {
		return fmt.Errorf("failed getting private key: %w", err)
	}

	if c.X509Certificate == "" {
		return c.generateX509Certificate(k, ca)
	}

	return nil
}
