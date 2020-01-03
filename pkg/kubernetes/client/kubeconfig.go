package client

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"text/template"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/pkg/types"
)

// Config is a simplified version of kubeconfig.
type Config struct {
	Server            string            `json:"server" yaml:"server"`
	CACertificate     types.Certificate `json:"caCertificate" yaml:"caCertificate"`
	ClientCertificate types.Certificate `json:"clientCertificate" yaml:"clientCertificate"`
	ClientKey         types.PrivateKey  `json:"clientKey" yaml:"clientKey"`
}

// Validate validates Config struct.
func (c *Config) Validate() error {
	if c.Server == "" {
		return fmt.Errorf("server is empty")
	}

	if c.ClientCertificate == "" {
		return fmt.Errorf("client certificate is empty")
	}

	if c.ClientKey == "" {
		return fmt.Errorf("client key is empty")
	}

	if c.CACertificate == "" {
		return fmt.Errorf("ca certificate is empty")
	}

	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config should succeed, got: %w", err)
	}

	if err := yaml.Unmarshal(b, c); err != nil {
		return fmt.Errorf("certificate validation failed: %w", err)
	}

	return nil
}

// ToYAMLString converts given configuration to kubeconfig format as YAML text
func (c *Config) ToYAMLString() (string, error) {
	if err := c.Validate(); err != nil {
		return "", fmt.Errorf("failed validating config: %w", err)
	}

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
    client-certificate-data: {{ .ClientCertificate }}
    client-key-data: {{ .ClientKey }}
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
	}{
		c.Server,
		base64.StdEncoding.EncodeToString([]byte(c.CACertificate)),
		base64.StdEncoding.EncodeToString([]byte(c.ClientCertificate)),
		base64.StdEncoding.EncodeToString([]byte(c.ClientKey)),
	}

	var buf bytes.Buffer

	tpl, err := template.New("t").Parse(t)
	if err != nil {
		return "", fmt.Errorf("failed parsing template: %w", err)
	}

	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed executing template: %w", err)
	}

	// Parse generated kubeconfig with Kubernetes client, to make sure everything is correct.
	if _, err := NewClient(buf.Bytes()); err != nil {
		return "", fmt.Errorf("generated kubeconfig is invalid: %w", err)
	}

	return buf.String(), nil
}
