package release_test

import (
	"testing"

	"github.com/flexkube/libflexkube/pkg/helm/release"
	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
	"github.com/flexkube/libflexkube/pkg/pki"
)

// New() tests.
//
//nolint:paralleltest // Helm client is not thread-safe.
func TestConfigNewBadKubeconfig(t *testing.T) {
	config := &release.Config{
		// Put content of your kubeconfig file here.
		Kubeconfig: "",

		// The namespace must be created upfront.
		Namespace: "kube-system",

		// Name of helm release.
		Name: "coredns",

		// Repositories must be added upfront as well.
		Chart: "stable/coredns",

		// Values passed to the release in YAML format.
		Values: `replicas: 1
labels:
  foo: bar
`,
		// Version of the chart to use.
		Version: "1.12.0",
	}

	if _, err := config.New(); err == nil {
		t.Fatalf("Creating release object with bad kubeconfig should fail")
	}
}

func newConfig(t *testing.T) *release.Config {
	t.Helper()

	pki := &pki.PKI{
		Certificate: pki.Certificate{
			RSABits: 512,
		},
		Kubernetes: &pki.Kubernetes{},
	}

	if err := pki.Generate(); err != nil {
		t.Fatalf("Generating PKI: %v", err)
	}

	clientConfig := client.Config{
		Server:            "foo",
		CACertificate:     pki.Kubernetes.CA.X509Certificate,
		ClientCertificate: pki.Kubernetes.AdminCertificate.X509Certificate,
		ClientKey:         pki.Kubernetes.AdminCertificate.PrivateKey,
	}

	kubeconfig, err := clientConfig.ToYAMLString()
	if err != nil {
		t.Fatalf("Rendering kubeconfig: %v", err)
	}

	return &release.Config{
		// Put content of your kubeconfig file here.
		Kubeconfig: kubeconfig,

		// The namespace must be created upfront.
		Namespace: "kube-system",

		// Name of helm release.
		Name: "coredns",

		// Repositories must be added upfront as well.
		Chart: "stable/asgasgasgsa",

		// Values passed to the release in YAML format.
		Values: `replicas: 1
labels:
  foo: bar
`,
		// Version of the chart to use.
		Version: "1.12.0",
	}
}

func newRelease(t *testing.T) release.Release {
	t.Helper()

	config := newConfig(t)

	r, err := config.New()
	if err != nil {
		t.Fatalf("Creating release object with valid kubeconfig should succeed: %v", err)
	}

	return r
}

//nolint:paralleltest // Helm client is not thread-safe.
func TestConfigNew(t *testing.T) {
	newRelease(t)
}

// Validate() tests.
//
//nolint:paralleltest // Helm client is not thread-safe.
func TestConfigValidateEmptyNamespace(t *testing.T) {
	c := newConfig(t)
	c.Namespace = ""

	if err := c.Validate(); err == nil {
		t.Fatalf("Validate should require namespace to be set")
	}
}

//nolint:paralleltest // Helm client is not thread-safe.
func TestConfigValidateEmptyName(t *testing.T) {
	c := newConfig(t)
	c.Name = ""

	if err := c.Validate(); err == nil {
		t.Fatalf("Validate should require name to be set")
	}
}

//nolint:paralleltest // Helm client is not thread-safe.
func TestConfigValidateEmptyChart(t *testing.T) {
	c := newConfig(t)
	c.Chart = ""

	if err := c.Validate(); err == nil {
		t.Fatalf("Validate should require chart to be set")
	}
}

//nolint:paralleltest // Helm client is not thread-safe.
func TestConfigValidateBadValues(t *testing.T) {
	c := newConfig(t)
	c.Values = "asd"

	if err := c.Validate(); err == nil {
		t.Fatalf("Validate should validate given values")
	}
}

// ValidateChart() tests.
//
//nolint:paralleltest // Helm client is not thread-safe.
func TestReleaseValidateChartBad(t *testing.T) {
	r := newRelease(t)

	if err := r.ValidateChart(); err == nil {
		t.Fatalf("Validating invalid chart should fail")
	}
}
