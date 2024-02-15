package client_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/internal/utiltest"
	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
	"github.com/flexkube/libflexkube/pkg/types"
)

// GetKubeconfig returns content of fake kubeconfig file for testing.
func GetKubeconfig(t *testing.T) string {
	t.Helper()

	pki := utiltest.GeneratePKI(t)

	testConfig := fmt.Sprintf(`server: %s
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

	clientConfig := &client.Config{}

	if err := yaml.Unmarshal([]byte(testConfig), clientConfig); err != nil {
		t.Fatalf("Unmarshaling config should succeed, got: %v", err)
	}

	kubeconfig, err := clientConfig.ToYAMLString()
	if err != nil {
		t.Fatalf("Generating kubeconfig should work, got: %v", err)
	}

	return kubeconfig
}

// ToYAMLString() tests.
func TestUnmarshal(t *testing.T) {
	t.Parallel()

	if kubeconfig := GetKubeconfig(t); kubeconfig == "" {
		t.Fatalf("Generated kubeconfig shouldn't be empty")
	}
}

func TestToYAMLStringNew(t *testing.T) { //nolint:funlen // Just many test cases.
	t.Parallel()

	cases := []struct {
		f   func(*client.Config)
		err func(*testing.T, error)
	}{
		{
			func(c *client.Config) {
				c.CACertificate = "ddd"
			},
			func(t *testing.T, err error) { //nolint:thelper // Actual test code.
				if err == nil {
					t.Errorf("Kubeconfig with bad CA Certificate should be invalid")
				}
			},
		},
		{
			func(c *client.Config) {
				c.ClientCertificate = "dfoo"
			},
			func(t *testing.T, err error) { //nolint:thelper // Actual test code.
				if err == nil {
					t.Errorf("Kubeconfig with bad client certificate should be invalid")
				}
			},
		},
		{
			func(c *client.Config) {
				c.ClientKey = "ffoo"
			},
			func(t *testing.T, err error) { //nolint:thelper // Actual test code.
				if err == nil {
					t.Errorf("Kubeconfig with bad client key should be invalid")
				}
			},
		},
		{
			func(c *client.Config) {
				pki := utiltest.GeneratePKI(t)
				c.ClientKey = types.PrivateKey(pki.PrivateKey)
			},
			func(t *testing.T, err error) { //nolint:thelper // Actual test code.
				if err == nil {
					t.Errorf("Kubeconfig with not matching client key should be invalid")
				}
			},
		},
		{
			func(*client.Config) {},
			func(t *testing.T, err error) { //nolint:thelper // Actual test code.
				if err != nil {
					t.Errorf("Valid config shouldn't return error, got: %v", err)
				}
			},
		},
		{
			func(c *client.Config) {
				c.ClientCertificate = ""
				c.ClientKey = ""
				c.Token = "doo"
			},
			func(t *testing.T, err error) { //nolint:thelper // Actual test code.
				if err != nil {
					t.Errorf("Config with only token set should be valid, got: %v", err)
				}
			},
		},
		{
			func(c *client.Config) {
				c.ClientCertificate = ""
				c.Token = "roo"
			},
			func(t *testing.T, err error) { //nolint:thelper // Actual test code.
				if err == nil {
					t.Errorf("Config with token and client key set should not be valid")
				}
			},
		},
		{
			func(c *client.Config) {
				c.ClientKey = ""
				c.Token = "fnoo"
			},
			func(t *testing.T, err error) { //nolint:thelper // Actual test code.
				if err == nil {
					t.Errorf("Config with token and client certificate set should be valid")
				}
			},
		},
	}

	for n, testCase := range cases {
		testCase := testCase

		t.Run(strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			pki := utiltest.GeneratePKI(t)

			config := &client.Config{
				Server:            "localhost",
				CACertificate:     types.Certificate(pki.Certificate),
				ClientCertificate: types.Certificate(pki.Certificate),
				ClientKey:         types.PrivateKey(pki.PrivateKey),
			}

			testCase.f(config)

			_, err := config.ToYAMLString()

			testCase.err(t, err)
		})
	}
}

func TestToYAMLStringValidate(t *testing.T) {
	t.Parallel()

	pki := utiltest.GeneratePKI(t)

	clientConfig := &client.Config{
		CACertificate:     types.Certificate(pki.Certificate),
		ClientCertificate: types.Certificate(pki.Certificate),
		ClientKey:         types.PrivateKey(pki.PrivateKey),
	}

	if _, err := clientConfig.ToYAMLString(); err == nil {
		t.Fatalf("ToYAMLString should validate the configuration")
	}
}

// Validate() tests.
func TestValidate(t *testing.T) { //nolint:funlen // There are just many test cases.
	t.Parallel()

	cases := []struct {
		f   func(*client.Config)
		err func(*testing.T, error)
	}{
		{
			func(c *client.Config) {
				c.Server = ""
			},
			func(t *testing.T, err error) { //nolint:thelper // Actual test code.
				if err == nil {
					t.Errorf("Kubeconfig without defined server should be invalid")
				}
			},
		},
		{
			func(c *client.Config) {
				c.CACertificate = ""
			},
			func(t *testing.T, err error) { //nolint:thelper // Actual test code.
				if err == nil {
					t.Errorf("Kubeconfig without defined CA Certificate should be invalid")
				}
			},
		},
		{
			func(c *client.Config) {
				c.ClientCertificate = ""
			},
			func(t *testing.T, err error) { //nolint:thelper // Actual test code.
				if err == nil {
					t.Errorf("Kubeconfig without defined client certificate should be invalid")
				}
			},
		},
		{
			func(c *client.Config) {
				c.ClientKey = ""
			},
			func(t *testing.T, err error) { //nolint:thelper // Actual test code.
				if err == nil {
					t.Errorf("Kubeconfig without defined client key should be invalid")
				}
			},
		},

		{
			func(*client.Config) {},
			func(t *testing.T, err error) { //nolint:thelper // Actual test code.
				if err != nil {
					t.Errorf("Valid config shouldn't return error, got: %v", err)
				}
			},
		},
	}

	for n, testCase := range cases {
		testCase := testCase

		t.Run(strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			pki := utiltest.GeneratePKI(t)

			config := &client.Config{
				Server:            "localhost",
				CACertificate:     types.Certificate(pki.Certificate),
				ClientCertificate: types.Certificate(pki.Certificate),
				ClientKey:         types.PrivateKey(pki.PrivateKey),
			}

			testCase.f(config)

			testCase.err(t, config.Validate())
		})
	}
}
