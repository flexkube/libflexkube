//go:build integration
// +build integration

package apiloadbalancer

import (
	"io/ioutil"
	"testing"

	"github.com/flexkube/libflexkube/internal/util"
)

func TestDeploy(t *testing.T) {
	t.Parallel()

	key, err := ioutil.ReadFile("/home/core/.ssh/id_rsa")
	if err != nil {
		t.Fatalf("reading SSH private key shouldn't fail, got: %v", err)
	}

	config := `
servers:
- 10.0.0.2
bindAddress: 0.0.0.0:7443
apiLoadBalancers:
- metricsBindAddress: 0.0.0.0
  host:
    ssh:
      address: localhost
ssh:
  user: core
  privateKey: |-
`
	config += util.Indent(string(key), "    ")

	c, err := FromYaml([]byte(config))
	if err != nil {
		t.Fatalf("creating apiloadbalancers object should succeed, got: %v", err)
	}

	if err := c.CheckCurrentState(); err != nil {
		t.Fatalf("checking current state should succeed, got: %v", err)
	}

	if err := c.Deploy(); err != nil {
		t.Fatalf("deploying should succeed, got: %v", err)
	}

	state, err := c.StateToYaml()
	if err != nil {
		t.Fatalf("dumping state should succeed, got: %v", err)
	}

	tearDownConfig := `
servers:
- 10.0.0.2
apiLoadBalancers: []
`

	c, err = FromYaml([]byte(tearDownConfig + string(state)))
	if err != nil {
		t.Fatalf("creating apiloadbalancers object for teardown should succeed, got: %v", err)
	}

	if err := c.CheckCurrentState(); err != nil {
		t.Fatalf("checking current state for teardown should succeed, got: %v", err)
	}

	if err := c.Deploy(); err != nil {
		t.Fatalf("tearing down should succeed, got: %v", err)
	}
}
