package release

import (
	"fmt"

	"helm.sh/helm/v3/pkg/kube"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/cmd/util"

	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
)

// TODO maybe we should return struct here?
func newClients(kubeconfig string) (*client.Getter, *kube.Client, *kubernetes.Clientset, error) {
	g, err := client.NewGetter([]byte(kubeconfig))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating Kubernetes client getter: %w", err)
	}

	c := &kube.Client{
		Factory: util.NewFactory(g),
		Log:     func(_ string, _ ...interface{}) {},
	}

	kc, err := c.Factory.KubernetesClientSet()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating Kubernetes client: %w", err)
	}

	return g, c, kc, nil
}
