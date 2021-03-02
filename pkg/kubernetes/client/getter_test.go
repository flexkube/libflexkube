package client_test

import (
	"testing"

	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
)

func TestGetter(t *testing.T) {
	t.Parallel()

	kubeconfig := GetKubeconfig(t)

	g, err := client.NewGetter([]byte(kubeconfig))
	if err != nil {
		t.Fatalf("Creating getter should work, got: %v", err)
	}

	if _, err := g.ToDiscoveryClient(); err != nil {
		t.Errorf("Turning getter into discovery client should work, got: %v", err)
	}

	if _, err := g.ToRESTMapper(); err != nil {
		t.Errorf("Turning getter into REST mapper should work, got: %v", err)
	}

	if c := g.ToRawKubeConfigLoader(); c == nil {
		t.Errorf("Turning getter into RawKubeConfigLoader should work")
	}
}
