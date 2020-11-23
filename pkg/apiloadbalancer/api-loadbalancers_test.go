package apiloadbalancer

import (
	"testing"

	"github.com/flexkube/libflexkube/pkg/types"
)

func TestPoolNoInstancesDefined(t *testing.T) {
	t.Parallel()

	a := &APILoadBalancers{}

	if err := a.Validate(); err == nil {
		t.Fatal("validate should fail if there is no instances defined and the state is empty")
	} else {
		t.Log(err)
	}
}

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

// New() tests.
func TestLoadBalancersNewValidate(t *testing.T) {
	t.Parallel()

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

// FromYaml() tests.
func TestLoadBalancersFromYaml(t *testing.T) {
	t.Parallel()
	GetLoadBalancers(t)
}

// StateToYaml() tests.
func TestLoadBalancersStateToYAML(t *testing.T) {
	t.Parallel()

	p := GetLoadBalancers(t)

	if _, err := p.StateToYaml(); err != nil {
		t.Fatalf("Dumping state to YAML should work, got: %v", err)
	}
}

// CheckCurrentState() tests.
func TestLoadBalancersCheckCurrentState(t *testing.T) {
	t.Parallel()

	p := GetLoadBalancers(t)

	if err := p.CheckCurrentState(); err != nil {
		t.Fatalf("Dumping state to YAML should work, got: %v", err)
	}
}

// Deploy() tests.
func TestLoadBalancersDeploy(t *testing.T) {
	t.Parallel()

	p := GetLoadBalancers(t)

	if err := p.Deploy(); err == nil {
		t.Fatalf("Deploying in testing environment should fail")
	}
}

// Containers() tests.
func TestLoadBalancersContainers(t *testing.T) {
	t.Parallel()

	p := GetLoadBalancers(t)

	if c := p.Containers(); c == nil {
		t.Fatalf("Containers() should return non-nil value")
	}
}
