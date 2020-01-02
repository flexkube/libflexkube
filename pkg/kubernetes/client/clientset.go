package client

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
)

// NewClientset returns Kubernetes clientset object from kubeconfig string.
func NewClientset(data []byte) (*kubernetes.Clientset, error) {
	cg, err := NewGetter(data)
	if err != nil {
		return nil, fmt.Errorf("failed creating kubernetes client getter: %w", err)
	}

	rc, err := cg.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed creating rest config: %w", err)
	}

	return kubernetes.NewForConfig(rc)
}
