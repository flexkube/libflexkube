//go:build integration
// +build integration

package apiloadbalancer

import (
	"os"
	"testing"

	"github.com/flexkube/libflexkube/internal/util"
)

//nolint:funlen,cyclop // Just lengthy test function.
func TestDeploy(t *testing.T) {
	t.Parallel()

	sshPrivateKeyPath := os.Getenv("TEST_INTEGRATION_SSH_PRIVATE_KEY_PATH")

	if sshPrivateKeyPath == "" {
		sshPrivateKeyPath = "/home/core/.ssh/id_rsa"
	}

	key, err := os.ReadFile(sshPrivateKeyPath)
	if err != nil {
		t.Fatalf("Reading SSH private key shouldn't fail, got: %v", err)
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

	if port := os.Getenv("TEST_INTEGRATION_SSH_PORT"); port != "" {
		config += "  port: " + port
	}

	loadBalancers, err := FromYaml([]byte(config))
	if err != nil {
		t.Fatalf("Creating apiloadbalancers object should succeed, got: %v", err)
	}

	if err := loadBalancers.CheckCurrentState(); err != nil {
		t.Fatalf("Checking current state should succeed, got: %v", err)
	}

	if err := loadBalancers.Deploy(); err != nil {
		t.Fatalf("Deploying should succeed, got: %v", err)
	}

	state, err := loadBalancers.StateToYaml()
	if err != nil {
		t.Fatalf("Dumping state should succeed, got: %v", err)
	}

	tearDownConfig := `
servers:
- 10.0.0.2
apiLoadBalancers: []
`

	loadBalancers, err = FromYaml([]byte(tearDownConfig + string(state)))
	if err != nil {
		t.Fatalf("Creating apiloadbalancers object for teardown should succeed, got: %v", err)
	}

	if err := loadBalancers.CheckCurrentState(); err != nil {
		t.Fatalf("Checking current state for teardown should succeed, got: %v", err)
	}

	if err := loadBalancers.Deploy(); err != nil {
		t.Fatalf("Tearing down should succeed, got: %v", err)
	}
}
