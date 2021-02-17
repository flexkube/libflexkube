package client

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"text/template"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/pkg/types"
)

// Config is a simplified version of kubeconfig.
type Config struct {
	// Server is a Kubernetes API server address.
	//
	// Example value: 'k8s.example.com:6443'.
	//
	// This field is required.
	Server string `json:"server,omitempty"`

	// CACertificate stores PEM encoded X.509 CA certificate, which was used
	// to sign Kubernetes API server certificate.
	//
	// This field is required.
	CACertificate types.Certificate `json:"caCertificate,omitempty"`

	// ClientCertificate stores PEM encoded X.509 client certificate, which will
	// be used for authentication and authorization to Kubernetes API server.
	//
	// This field is optional if Token field is populated.
	ClientCertificate types.Certificate `json:"clientCertificate,omitempty"`

	// ClientCertificate stores PEM encoded private key in PKCS1, PKCS8 or EC formats,
	// which will be used for authentication and authorization to Kubernetes API server.
	// Key must match configured ClientCertificate.
	//
	// This field is optional if Token field is populated.
	ClientKey types.PrivateKey `json:"clientKey,omitempty"`

	// Token stores Kubernetes token, which will be used for authentication and authrization
	// to Kubernetes API server. Usually used by kubelet to perform TLS bootstrapping.
	Token string `json:"token,omitempty"`
}

// Validate validates Config struct.
func (c *Config) Validate() error {
	var errors util.ValidateError

	if c.Server == "" {
		errors = append(errors, fmt.Errorf("server is empty"))
	}

	if c.CACertificate == "" {
		errors = append(errors, fmt.Errorf("ca certificate is empty"))
	}

	errors = append(errors, c.validateAuth()...)

	b, err := yaml.Marshal(c)
	if err != nil {
		return append(errors, fmt.Errorf("marshaling config should succeed, got: %w", err))
	}

	if err := yaml.Unmarshal(b, c); err != nil {
		return append(errors, fmt.Errorf("certificate validation failed: %w", err))
	}

	return errors.Return()
}

func (c *Config) validateAuth() util.ValidateError {
	var errors util.ValidateError

	if c.ClientCertificate == "" && c.Token == "" {
		errors = append(errors, fmt.Errorf("either client certificate or token must be set"))
	}

	if c.ClientKey == "" && c.Token == "" {
		errors = append(errors, fmt.Errorf("either client key or token must be set"))
	}

	if c.Token != "" && c.ClientCertificate != "" {
		errors = append(errors, fmt.Errorf("client certificate should not be set together with token"))
	}

	if c.Token != "" && c.ClientKey != "" {
		errors = append(errors, fmt.Errorf("client key should not be set together with token"))
	}

	return errors
}

// ToYAMLString converts given configuration to kubeconfig format as YAML text.
func (c *Config) ToYAMLString() (string, error) {
	if err := c.Validate(); err != nil {
		return "", fmt.Errorf("failed validating config: %w", err)
	}

	kubeconfig, err := c.renderKubeconfig()
	if err != nil {
		return "", fmt.Errorf("failed rendering kubeconfig: %w", err)
	}

	// Parse generated kubeconfig with Kubernetes client, to make sure everything is correct.
	if _, err := NewClient([]byte(kubeconfig)); err != nil {
		return "", fmt.Errorf("generated kubeconfig is invalid: %w", err)
	}

	return kubeconfig, nil
}

// renderKubeconfig renders Config as kubeconfig YAML.
func (c *Config) renderKubeconfig() (string, error) {
	t := `apiVersion: v1
kind: Config
clusters:
- name: static
  cluster:
    server: https://{{ .Server }}
    certificate-authority-data: {{ .CACertificate }}
users:
- name: static
  user:
    {{- if .ClientCertificate }}
    client-certificate-data: {{ .ClientCertificate }}
    {{- end }}
    {{- if .ClientKey }}
    client-key-data: {{ .ClientKey }}
    {{- end }}
    {{- if .Token }}
    token: {{ .Token }}
    {{- end }}
current-context: static
contexts:
- name: static
  context:
    cluster: static
    user: static
`

	data := struct {
		Server            string
		CACertificate     string
		ClientCertificate string
		ClientKey         string
		Token             string
	}{
		c.Server,
		base64.StdEncoding.EncodeToString([]byte(c.CACertificate)),
		base64.StdEncoding.EncodeToString([]byte(c.ClientCertificate)),
		base64.StdEncoding.EncodeToString([]byte(c.ClientKey)),
		c.Token,
	}

	var buf bytes.Buffer

	tpl := template.Must(template.New("t").Parse(t))

	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed executing template: %w", err)
	}

	return buf.String(), nil
}
