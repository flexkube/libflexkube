package kubelet

import (
	"testing"
)

func GetPool(t *testing.T) *pool {
	y := `
ssh: {}
bootstrapKubeconfig: foo
kubelets:
- {}
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
