package client

import (
	"fmt"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/internal/utiltest"
	"github.com/flexkube/libflexkube/pkg/types"
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
func TestUnmarshal(t *testing.T) {
	if kubeconfig := GetKubeconfig(t); kubeconfig == "" {
		t.Fatalf("Generated kubeconfig shouldn't be empty")
	}
}

func TestToYAMLStringNew(t *testing.T) {
	cases := []struct {
		f   func(*Config)
		err func(error, *testing.T)
	}{
		{
			func(c *Config) {
				c.CACertificate = "ddd"
			},
			func(err error, t *testing.T) {
				if err == nil {
					t.Errorf("Kubeconfig with bad CA Certificate should be invalid")
				}
			},
		},
		{
			func(c *Config) {
				c.ClientCertificate = "foo"
			},
			func(err error, t *testing.T) {
				if err == nil {
					t.Errorf("Kubeconfig with bad client certificate should be invalid")
				}
			},
		},
		{
			func(c *Config) {
				c.ClientKey = "foo"
			},
			func(err error, t *testing.T) {
				if err == nil {
					t.Errorf("Kubeconfig with bad client key should be invalid")
				}
			},
		},
		{
			func(c *Config) {
				pki := utiltest.GeneratePKI(t)
				c.ClientKey = types.PrivateKey(pki.PrivateKey)
			},
			func(err error, t *testing.T) {
				if err == nil {
					t.Errorf("Kubeconfig with not matching client key should be invalid")
				}
			},
		},

		{
			func(c *Config) {},
			func(err error, t *testing.T) {
				if err != nil {
					t.Errorf("Valid config shouldn't return error, got: %v", err)
				}
			},
		},
	}

	for n, c := range cases {
		c := c

		t.Run(fmt.Sprintf("%d", n), func(t *testing.T) {
			t.Parallel()

			pki := utiltest.GeneratePKI(t)

			config := &Config{
				Server:            "localhost",
				CACertificate:     types.Certificate(pki.Certificate),
				ClientCertificate: types.Certificate(pki.Certificate),
				ClientKey:         types.PrivateKey(pki.PrivateKey),
			}

			c.f(config)

			_, err := config.ToYAMLString()

			c.err(err, t)
		})
	}
}

func TestToYAMLStringValidate(t *testing.T) {
	pki := utiltest.GeneratePKI(t)

	c := &Config{
		CACertificate:     types.Certificate(pki.Certificate),
		ClientCertificate: types.Certificate(pki.Certificate),
		ClientKey:         types.PrivateKey(pki.PrivateKey),
	}

	if _, err := c.ToYAMLString(); err == nil {
		t.Fatalf("ToYAMLString should validate the configuration")
	}
}

// Validate()
func TestValidate(t *testing.T) {
	cases := []struct {
		f   func(*Config)
		err func(error, *testing.T)
	}{
		{
			func(c *Config) {
				c.Server = ""
			},
			func(err error, t *testing.T) {
				if err == nil {
					t.Errorf("Kubeconfig without defined server should be invalid")
				}
			},
		},
		{
			func(c *Config) {
				c.CACertificate = ""
			},
			func(err error, t *testing.T) {
				if err == nil {
					t.Errorf("Kubeconfig without defined CA Certificate should be invalid")
				}
			},
		},
		{
			func(c *Config) {
				c.ClientCertificate = ""
			},
			func(err error, t *testing.T) {
				if err == nil {
					t.Errorf("Kubeconfig without defined client certificate should be invalid")
				}
			},
		},
		{
			func(c *Config) {
				c.ClientKey = ""
			},
			func(err error, t *testing.T) {
				if err == nil {
					t.Errorf("Kubeconfig without defined client key should be invalid")
				}
			},
		},

		{
			func(c *Config) {},
			func(err error, t *testing.T) {
				if err != nil {
					t.Errorf("Valid config shouldn't return error, got: %v", err)
				}
			},
		},
	}

	for n, c := range cases {
		c := c

		t.Run(fmt.Sprintf("%d", n), func(t *testing.T) {
			t.Parallel()

			pki := utiltest.GeneratePKI(t)

			config := &Config{
				Server:            "localhost",
				CACertificate:     types.Certificate(pki.Certificate),
				ClientCertificate: types.Certificate(pki.Certificate),
				ClientKey:         types.PrivateKey(pki.PrivateKey),
			}

			c.f(config)

			c.err(config.Validate(), t)
		})
	}
}
