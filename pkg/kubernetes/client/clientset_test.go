package client_test

import (
	"testing"

	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
)

func TestNewClientset(t *testing.T) {
	t.Parallel()

	kubeconfig := GetKubeconfig(t)

	if _, err := client.NewClientset([]byte(kubeconfig)); err != nil {
		t.Fatalf("Creating clientset should work, got: %v", err)
	}
}
