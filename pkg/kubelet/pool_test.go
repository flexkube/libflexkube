package kubelet

import (
	"testing"

	"github.com/flexkube/libflexkube/pkg/types"
)

func GetPool(t *testing.T) types.Resource {
	y := `
ssh:
  address: localhost
  password: foo
  connectionTimeout: 1s
  retryTimeout: 1s
  retryInterval: 1s
bootstrapKubeconfig: foo
volumePluginDir: /var/lib/kubelet/volumeplugins
kubelets:
- networkPlugin: cni
`

	p, err := FromYaml([]byte(y))
	if err != nil {
		t.Fatalf("Creating pool from YAML should succeed, got: %v", err)
	}

	return p
}

// New()
func TestPoolNewValidate(t *testing.T) {
	y := `
ssh:
  address: localhost
  password: foo
  connectionTimeout: 1s
  retryTimeout: 1s
  retryInterval: 1s
volumePluginDir: /var/lib/kubelet/volumeplugins
kubelets:
- networkPlugin: cni
`

	if _, err := FromYaml([]byte(y)); err == nil {
		t.Fatalf("Creating pool from bad YAML should fail")
	}
}

// FromYaml()
func TestPoolFromYaml(t *testing.T) {
	GetPool(t)
}

// StateToYaml()
func TestPoolStateToYAML(t *testing.T) {
	p := GetPool(t)

	if _, err := p.StateToYaml(); err != nil {
		t.Fatalf("Dumping state to YAML should work, got: %v", err)
	}
}

// CheckCurrentState()
func TestPoolCheckCurrentState(t *testing.T) {
	p := GetPool(t)

	if err := p.CheckCurrentState(); err != nil {
		t.Fatalf("Checking current state of empty pool should work, got: %v", err)
	}
}

// Containers()
func TestPoolContainers(t *testing.T) {
	p := GetPool(t)

	if c := p.Containers(); c == nil {
		t.Fatalf("Containers() should return non-nil value")
	}
}

// Deploy()
func TestPoolDeploy(t *testing.T) {
	p := GetPool(t)

	if err := p.Deploy(); err == nil {
		t.Fatalf("Deploying in testing environment should fail")
	}
}
