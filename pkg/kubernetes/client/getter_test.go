package client

import (
	"testing"
)

func TestGetter(t *testing.T) {
	kubeconfig := GetKubeconfig(t)

	g, err := NewGetter([]byte(kubeconfig))
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
