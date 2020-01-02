package client

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"text/template"
)

// Config is a simlified version of kubeconfig.
type Config struct {
	Server            string `json:"server" yaml:"server"`
	CACertificate     string `json:"caCertificate" yaml:"caCertificate"`
	ClientCertificate string `json:"clientCertificate" yaml:"clientCertificate"`
	ClientKey         string `json:"clientKey" yaml:"clientKey"`
}

// ToYAMLString converts given configuration to kubeconfig format as YAML text
func (c *Config) ToYAMLString() (string, error) {
	t := `apiVersion: v1
kind: Config
clusters:
- name: static
  cluster:
    server: https://{{ .Server }}:6443
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
