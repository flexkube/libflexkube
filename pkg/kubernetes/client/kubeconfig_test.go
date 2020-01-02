package client

import (
	"fmt"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/internal/utiltest"
)

// GetKubeconfig returns content of fake kubeconfig file for testing.
func GetKubeconfig(t *testing.T) string {
	pki := utiltest.GeneratePKI(t)

	y := fmt.Sprintf(`server: %s
caCertificate: |
  %s
clientCertificate: |
  %s
clientKey: |
  %s
`,
		"localhost",
		strings.TrimSpace(util.Indent(pki.Certificate, "  ")),
		strings.TrimSpace(util.Indent(pki.Certificate, "  ")),
		strings.TrimSpace(util.Indent(pki.PrivateKey, "  ")),
	)

	c := &Config{}

	if err := yaml.Unmarshal([]byte(y), c); err != nil {
		t.Fatalf("unmarshaling config should succeed, got: %v", err)
	}

	kubeconfig, err := c.ToYAMLString()
	if err != nil {
		t.Fatalf("Generating kubeconfig should work, got: %v", err)
	}

	return kubeconfig
}

// ToYAMLString()
func TestToYAMLString(t *testing.T) {
	if kubeconfig := GetKubeconfig(t); kubeconfig == "" {
		t.Fatalf("Generated kubeconfig shouldn't be empty")
	}
}
