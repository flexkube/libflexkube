package release

import (
	"fmt"

	"github.com/flexkube/helm/v3/pkg/kube"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/cmd/util"

	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
)

// TODO maybe we should return struct here?
func newClients(kubeconfig string) (*client.Getter, *kube.Client, *kubernetes.Clientset, error) {
	clientGetter, err := client.NewGetter([]byte(kubeconfig))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating Kubernetes client getter: %w", err)
	}

	client := &kube.Client{
		Factory: util.NewFactory(clientGetter),
		Log:     func(_ string, _ ...interface{}) {},
	}

	kc, err := client.Factory.KubernetesClientSet()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating Kubernetes client: %w", err)
	}

	return clientGetter, client, kc, nil
}
