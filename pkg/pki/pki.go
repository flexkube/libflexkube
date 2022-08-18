// Package pki allows to manage Kubernetes PKI certificates.
package pki

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"sort"
	"strings"
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

func keyUsageFromString(usageRaw string) x509.KeyUsage {
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
	}[usageRaw]
}

func extendedKeyUsageFromString(usageRaw string) x509.ExtKeyUsage {
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
	}[usageRaw]
}

// Certificate defines configurable options for each certificate.
type Certificate struct {
	// Organization stores value for 'organization' field in the certificate.
	Organization string `json:"organization,omitempty"`

	// RSABits defines length of RSA private key to generate.
	//
	// Example value: '2048'.
	RSABits int `json:"rsaBits,omitempty"`

	// ValidityDuration defines how long generated certificates should be valid.
	//
	// Example value: '24h'.
	ValidityDuration string `json:"validityDuration,omitempty"`

	// RenewThreshold defines how long before expiry date the certificates should
	// be re-generated.
	RenewThreshold string `json:"renewThreshold,omitempty"`

	// CommonName defined CN field for the certificate.
	CommonName string `json:"commonName,omitempty"`

	// CA controls if certificate should be self-signed while generated.
	CA bool `json:"ca,omitempty"`

	// KeyUsage is a list of key usages. Valid values are:
	// - "digital_signature"
	// - "content_commitment"
	// - "key_encipherment"
	// - "data_encipherment"
	// - "key_agreement"
	// - "cert_signing"
	// - "crl_signing"
	// - "encipher_only"
	// - "decipher_only"
	// - "any_extended"
	// - "server_auth"
	// - "client_auth"
	// - "code_signing"
	// - "email_protection"
	// - "ipsec_end_system"
	// - "ipsec_tunnel"
	// - "ipsec_user"
	// - "timestamping"
	// - "ocsp_signing"
	// - "microsoft_server_gated_crypto"
	// - "netscape_server_gated_crypto"
	KeyUsage []string `json:"keyUsage,omitempty"`

	// IPAddresses defines for which IP addresses the certificate can be used.
	IPAddresses []string `json:"ipAddresses,omitempty"`

	// DNSNames defines extra hostnames, which will be valid for the certificate.
	DNSNames []string `json:"dnsNames,omitempty"`

	// X509Certificate stores generated certificate in X.509 certificate format, PEM encoded.
	X509Certificate types.Certificate `json:"x509Certificate,omitempty"`

	// PublicKey stores generate RSA public key, PEM encoded.
	PublicKey string `json:"publicKey,omitempty"`

	// PrivateKey stores generates RSA private key in PKCS1 format, PEM encoded.
	PrivateKey types.PrivateKey `json:"privateKey,omitempty"`
}

// PKI contains configuration and all generated certificates and private keys required for running Kubernetes.
type PKI struct {
	// Certificate contains default settings for all certificates in PKI.
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
	for _, certRequest := range crs {
		cert, err := buildCertificate(certRequest.Certificates...)
		if err != nil {
			return fmt.Errorf("building certificate configuration: %w", err)
		}

		if err := cert.Generate(certRequest.CA); err != nil {
			return fmt.Errorf("generating the certificate: %w", err)
		}

		if certRequest.Target == nil {
			return fmt.Errorf("target certificate is not set")
		}

		certRequest.Target.X509Certificate = cert.X509Certificate
		certRequest.Target.PrivateKey = cert.PrivateKey
		certRequest.Target.PublicKey = cert.PublicKey
	}

	return nil
}

func (p *PKI) generateRootCA() error {
	if p.RootCA == nil {
		p.RootCA = &Certificate{}
	}

	certRequest := &certificateRequest{
		Target: p.RootCA,
		Certificates: []*Certificate{
			&p.Certificate,
			caCertificate(RootCACN),
			p.RootCA,
		},
	}

	if err := buildAndGenerate(certRequest); err != nil {
		return fmt.Errorf("generating root CA certificate: %w", err)
	}

	return nil
}

// Generate generates PKI required for running Kubernetes, including root CA and etcd certificates.
func (p *PKI) Generate() error {
	if err := p.generateRootCA(); err != nil {
		return fmt.Errorf("generating root CA certificate: %w", err)
	}

	// If etcd field is set, generate etcd PKI. This allows to skip generation of those certificates,
	// if one deploys just Kubernetes on existing etcd cluster.
	if p.Etcd != nil {
		if err := p.Etcd.Generate(p.RootCA, p.Certificate); err != nil {
			return fmt.Errorf("generating etcd PKI: %w", err)
		}
	}

	// If Kubernetes field is set, generate Kubernetes PKI. This allows to skip generation of those certificates,
	// if one deploys just etcd cluster.
	if p.Kubernetes != nil {
		if err := p.Kubernetes.Generate(p.RootCA, p.Certificate); err != nil {
			return fmt.Errorf("generating Kubernetes PKI: %w", err)
		}
	}

	return nil
}

// buildCertificate merges N number of given certificates. Properties of last given certificate takes
// precedence over previous ones.
func buildCertificate(certs ...*Certificate) (*Certificate, error) {
	cert := &Certificate{
		Organization:     Organization,
		RSABits:          RSABits,
		ValidityDuration: ValidityDuration,
		RenewThreshold:   RenewThreshold,
	}

	for _, c := range certs {
		rc, err := yaml.Marshal(c)
		if err != nil {
			return nil, fmt.Errorf("marshaling certificate: %w", err)
		}

		if err := yaml.Unmarshal(rc, cert); err != nil {
			return nil, fmt.Errorf("unmarshaling the certificate: %w", err)
		}
	}

	return cert, nil
}

func (c *Certificate) decodePrivateKey() (*rsa.PrivateKey, error) {
	der, _ := pem.Decode([]byte(c.PrivateKey))
	if der == nil {
		return nil, fmt.Errorf("private key is not defined in valid PEM format")
	}

	k, err := x509.ParsePKCS1PrivateKey(der.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing private key to PKCS1 format: %w", err)
	}

	return k, nil
}

// DecodeX509Certificate returns parsed version of X.509 certificate, so one can read
// the fields of generated certificate.
func (c *Certificate) DecodeX509Certificate() (*x509.Certificate, error) {
	der, _ := pem.Decode([]byte(c.X509Certificate))
	if der == nil {
		//nolint:stylecheck // Capitaliziation is OK here, as X.509 is a proper noun.
		return nil, fmt.Errorf("X.509 certificate is not defined in valid PEM format")
	}

	cert, err := x509.ParseCertificate(der.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing X.509 certificate: %w", err)
	}

	return cert, nil
}

// persistPublicKey persist given RSA public key into the certificate object.
func (c *Certificate) persistPublicKey(k interface{}) error {
	pubBytes, err := x509.MarshalPKIXPublicKey(k)
	if err != nil {
		return fmt.Errorf("marshaling RSA public key: %w", err)
	}

	var buf bytes.Buffer

	if err := pem.Encode(&buf, &pem.Block{Type: RSAPublicKeyPEMHeader, Bytes: pubBytes}); err != nil {
		return fmt.Errorf("encoding RSA public key: %w", err)
	}

	c.PublicKey = buf.String()

	return nil
}

func (c *Certificate) generatePrivateKey() (*rsa.PrivateKey, error) {
	// generate RSA private key.
	privateKey, err := rsa.GenerateKey(rand.Reader, c.RSABits)
	if err != nil {
		return nil, fmt.Errorf("generating RSA key: %w", err)
	}

	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	var buf bytes.Buffer
	if err := pem.Encode(&buf, &pem.Block{Type: RSAPrivateKeyPEMHeader, Bytes: privBytes}); err != nil {
		return nil, fmt.Errorf("encoding RSA private key: %w", err)
	}

	c.PrivateKey = types.PrivateKey(buf.String())

	if err := c.persistPublicKey(privateKey.Public()); err != nil {
		return nil, fmt.Errorf("persisting RSA public key: %w", err)
	}

	return privateKey, nil
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
		return fmt.Errorf("parsing validity duration %q for certificate: %w", c.ValidityDuration, err)
	}

	for _, i := range c.IPAddresses {
		if ip := net.ParseIP(i); ip == nil {
			return fmt.Errorf("parsing IP address %q", i)
		}
	}

	if c.RSABits == 0 {
		return fmt.Errorf("RSA bits can't be 0")
	}

	return nil
}

func (c *Certificate) decodeKeyUsage() (x509.KeyUsage, []x509.ExtKeyUsage) {
	keyUsage := 0
	extendedKeyUsage := []x509.ExtKeyUsage{}

	for _, rawKeyUsage := range c.KeyUsage {
		r := int(keyUsageFromString(rawKeyUsage))
		if r != 0 {
			keyUsage |= r

			continue
		}

		if e := extendedKeyUsageFromString(rawKeyUsage); e != 0 {
			extendedKeyUsage = append(extendedKeyUsage, e)
		}
	}

	return x509.KeyUsage(keyUsage), extendedKeyUsage
}

func (c *Certificate) generateX509Certificate(certPK *rsa.PrivateKey, caCert *Certificate) error {
	// Generate serial number for X.509 certificate.
	//
	//nolint:gomnd // As in https://golang.org/src/crypto/tls/generate_cert.go.
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("generating serial number for certificate: %w", err)
	}

	validityDuration, _ := time.ParseDuration(c.ValidityDuration) //nolint:errcheck // Already done in Validate().

	keyUsage, extendedKeyUsage := c.decodeKeyUsage()

	cert := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{c.Organization},
			CommonName:   c.CommonName,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(validityDuration),

		KeyUsage:              keyUsage,
		ExtKeyUsage:           extendedKeyUsage,
		BasicConstraintsValid: true,
		IsCA:                  c.CA,
		DNSNames:              c.DNSNames,
	}

	for _, i := range c.IPAddresses {
		cert.IPAddresses = append(cert.IPAddresses, net.ParseIP(i))
	}

	caPK := certPK
	x509CACert := &cert

	if caCert != nil {
		x509CACert, caPK, err = caCert.decodeKeypair()
		if err != nil {
			return fmt.Errorf("decoding CA key pair: %w", err)
		}
	}

	subjectKeyID, err := bigIntHash(certPK.N)
	if err != nil {
		return fmt.Errorf("generating certificate subject Key ID: %w", err)
	}

	cert.SubjectKeyId = subjectKeyID

	return c.createAndPersist(&cert, x509CACert, certPK, caPK)
}

func (c *Certificate) createAndPersist(cert, caCert *x509.Certificate, certPK, caPK *rsa.PrivateKey) error {
	der, err := x509.CreateCertificate(rand.Reader, cert, caCert, &certPK.PublicKey, caPK)
	if err != nil {
		return fmt.Errorf("creating certificate: %w", err)
	}

	return c.persistX509Certificate(der)
}

// Taken from https://play.golang.org/p/tispiUVmdm.
func bigIntHash(n *big.Int) ([]byte, error) {
	hash := sha1.New() // #nosec G401

	if _, err := hash.Write(n.Bytes()); err != nil {
		return nil, fmt.Errorf("writing bytes to SHA1 function: %w", err)
	}

	return hash.Sum(nil), nil
}

// decodeKeypair decodes both X.509 certificate and private key.
func (c *Certificate) decodeKeypair() (*x509.Certificate, *rsa.PrivateKey, error) {
	privateKey, err := c.decodePrivateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("decoding private key: %w", err)
	}

	cert, err := c.DecodeX509Certificate()
	if err != nil {
		return nil, nil, fmt.Errorf("decoding X.509 certificate: %w", err)
	}

	return cert, privateKey, nil
}

func (c *Certificate) persistX509Certificate(der []byte) error {
	// Encode private certificate from PEM into PEM format.
	var cert bytes.Buffer

	if err := pem.Encode(&cert, &pem.Block{Type: X509CertificatePEMHeader, Bytes: der}); err != nil {
		return fmt.Errorf("writing data to cert.pem: %w", err)
	}

	c.X509Certificate = types.Certificate(cert.String())

	return nil
}

// Generate ensures that all fields of the certificate are populated.
//
// This function currently supports:
//
// - Generating new RSA private key and public key.
//
// - Generating new X.509 certificates.
//
// - Re-generating X.509 certificate if IP addresses changes.
//
// NOT implemented functionality:
//
// - Renewing certificates based on expiry time.
//
// - Renewing X.509 certificate after RSA private key renewal.
//
// - Renewing issued certificate during CA renewal.
func (c *Certificate) Generate(caCert *Certificate) error {
	if err := c.Validate(); err != nil {
		return fmt.Errorf("validating the certificate: %w", err)
	}

	k, err := c.getPrivateKey()
	if err != nil {
		return fmt.Errorf("getting private key: %w", err)
	}

	return c.ensureX509Certificate(k, caCert)
}

// ensureX509Certificate checks if the certificate is up to date and if not, triggers
// certificate generation.
func (c *Certificate) ensureX509Certificate(privateKey *rsa.PrivateKey, caCert *Certificate) error {
	upToDate, err := c.IsX509CertificateUpToDate()
	if err != nil {
		return fmt.Errorf("checking if X.509 certificate is up to date: %w", err)
	}

	if !upToDate {
		return c.generateX509Certificate(privateKey, caCert)
	}

	return nil
}

func ipAddressesUpToDate(cert *x509.Certificate, configuredIPs []string) bool {
	ips := []string{}

	for _, i := range cert.IPAddresses {
		ips = append(ips, i.String())
	}

	sort.Strings(ips)

	sort.Strings(configuredIPs)

	return strings.Join(ips, ",") == strings.Join(configuredIPs, ",")
}

// IsX509CertificateUpToDate checks, if generated X.509 certificate is up to date
// with it's configuration.
func (c *Certificate) IsX509CertificateUpToDate() (bool, error) {
	if c.X509Certificate == "" {
		return false, nil
	}

	cert, err := c.DecodeX509Certificate()
	if err != nil {
		return true, fmt.Errorf("decoding X.509 certificate: %w", err)
	}

	if !ipAddressesUpToDate(cert, c.IPAddresses) {
		return false, nil
	}

	return true, nil
}
