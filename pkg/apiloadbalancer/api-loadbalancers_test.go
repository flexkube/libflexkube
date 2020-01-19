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
apiLoadBalancers:
- metricsBindAddress: 0.0.0.0:2222
servers:
- localhost:6443
`

	p, err := FromYaml([]byte(y))
	if err != nil {
		t.Fatalf("Creating load balancers from YAML should succeed, got: %v", err)
	}

	return p
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
