package release_test

import (
	"testing"

	"github.com/flexkube/libflexkube/pkg/helm/release"
	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
	"github.com/flexkube/libflexkube/pkg/pki"
)

// New() tests.
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
		t.Fatalf("creating release object with bad kubeconfig should fail")
	}
}

func newConfig(t *testing.T) *release.Config {
	pki := &pki.PKI{
		Certificate: pki.Certificate{
			RSABits: 512,
		},
		Kubernetes: &pki.Kubernetes{},
	}

	if err := pki.Generate(); err != nil {
		t.Fatalf("generating PKI: %v", err)
	}

	c := client.Config{
		Server:            "foo",
		CACertificate:     pki.Kubernetes.CA.X509Certificate,
		ClientCertificate: pki.Kubernetes.AdminCertificate.X509Certificate,
		ClientKey:         pki.Kubernetes.AdminCertificate.PrivateKey,
	}

	kubeconfig, err := c.ToYAMLString()
	if err != nil {
		t.Fatalf("rendering kubeconfig: %v", err)
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
	config := newConfig(t)

	r, err := config.New()
	if err != nil {
		t.Fatalf("creating release object with valid kubeconfig should succeed: %v", err)
	}

	return r
}

func TestConfigNew(t *testing.T) {
	t.Parallel()

	newRelease(t)
}

// Validate() tests.
func TestConfigValidateEmptyNamespace(t *testing.T) {
	t.Parallel()

	c := newConfig(t)
	c.Namespace = ""

	if err := c.Validate(); err == nil {
		t.Fatalf("validate should require namespace to be set")
	}
}

func TestConfigValidateEmptyName(t *testing.T) {
	t.Parallel()

	c := newConfig(t)
	c.Name = ""

	if err := c.Validate(); err == nil {
		t.Fatalf("validate should require name to be set")
	}
}

func TestConfigValidateEmptyChart(t *testing.T) {
	t.Parallel()

	c := newConfig(t)
	c.Chart = ""

	if err := c.Validate(); err == nil {
		t.Fatalf("validate should require chart to be set")
	}
}

func TestConfigValidateBadValues(t *testing.T) {
	t.Parallel()

	c := newConfig(t)
	c.Values = "asd"

	if err := c.Validate(); err == nil {
		t.Fatalf("validate should validate given values")
	}
}

// ValidateChart() tests.
func TestReleaseValidateChartBad(t *testing.T) {
	t.Parallel()

	r := newRelease(t)

	if err := r.ValidateChart(); err == nil {
		t.Fatalf("validating invalid chart should fail")
	}
}
