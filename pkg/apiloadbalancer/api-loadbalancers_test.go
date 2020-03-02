package apiloadbalancer

import (
	"testing"

	"github.com/flexkube/libflexkube/pkg/types"
)

func GetLoadBalancers(t *testing.T) types.Resource {
	y := `
ssh:
  address: localhost
  password: foo
  connectionTimeout: 1s
  retryTimeout: 1s
  retryInterval: 1s
bindAddress: 0.0.0.0:6443
apiLoadBalancers:
- {}
servers:
- localhost:6443
`

	p, err := FromYaml([]byte(y))
	if err != nil {
		t.Fatalf("Creating load balancers from YAML should succeed, got: %v", err)
	}

	return p
}

// New()
func TestLoadBalancersNewValidate(t *testing.T) {
	y := `
ssh:
  address: localhost
  password: foo
  connectionTimeout: 1s
  retryTimeout: 1s
  retryInterval: 1s
bindAddress: 0.0.0.0:6443
apiLoadBalancers:
- {}
`

	if _, err := FromYaml([]byte(y)); err == nil {
		t.Fatalf("Creating load balancers from bad YAML should fail")
	}
}

// FromYaml()
func TestLoadBalancersFromYaml(t *testing.T) {
	GetLoadBalancers(t)
}

// StateToYaml()
func TestLoadBalancersStateToYAML(t *testing.T) {
	p := GetLoadBalancers(t)

	if _, err := p.StateToYaml(); err != nil {
		t.Fatalf("Dumping state to YAML should work, got: %v", err)
	}
}

// CheckCurrentState()
func TestLoadBalancersCheckCurrentState(t *testing.T) {
	p := GetLoadBalancers(t)

	if err := p.CheckCurrentState(); err != nil {
		t.Fatalf("Dumping state to YAML should work, got: %v", err)
	}
}

// Deploy()
func TestLoadBalancersDeploy(t *testing.T) {
	p := GetLoadBalancers(t)

	if err := p.Deploy(); err == nil {
		t.Fatalf("Deploying in testing environment should fail")
	}
}

// Containers()
func TestLoadBalancersContainers(t *testing.T) {
	p := GetLoadBalancers(t)

	if c := p.Containers(); c == nil {
		t.Fatalf("Containers() should return non-nil value")
	}
}
