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
kubelets:
- networkPlugin: cni
`

	p, err := FromYaml([]byte(y))
	if err != nil {
		t.Fatalf("Creating pool from YAML should succeed, got: %v", err)
	}

	return p
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
		t.Fatalf("Dumping state to YAML should work, got: %v", err)
	}
}

// Deploy()
func TestPoolDeploy(t *testing.T) {
	p := GetPool(t)

	if err := p.Deploy(); err == nil {
		t.Fatalf("Deploying in testing environment should fail")
	}
}
