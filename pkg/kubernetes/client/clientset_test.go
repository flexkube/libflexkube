package client

import (
	"testing"
)

func TestNewClientset(t *testing.T) {
	kubeconfig := GetKubeconfig(t)

	if _, err := NewClientset([]byte(kubeconfig)); err != nil {
		t.Fatalf("Creating clientset should work, got: %v", err)
	}
}
