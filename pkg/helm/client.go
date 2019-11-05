package helm

import (
	"fmt"

	"helm.sh/helm/v3/pkg/kube"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kubectl/pkg/cmd/util"

	"github.com/invidian/libflexkube/pkg/kubernetes/client"
)

// TODO maybe we should return struct here?
func newClients(kubeconfig string) (*client.Getter, *kube.Client, *kubernetes.Clientset, error) {

	// Inlining helm.sh/helm/v3/pkg/kube.New() to be able to override the config
	if err := v1beta1.AddToScheme(scheme.Scheme); err != nil {
		// According to helm, this error should never happen
		return nil, nil, nil, fmt.Errorf("unknown error occurred: %w", err)
	}

	g, err := client.NewGetter([]byte(kubeconfig))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create kubernetes client getter: %w", err)
	}

	c := &kube.Client{
		Factory: util.NewFactory(g),
		Log:     func(_ string, _ ...interface{}) {},
	}

	kc, err := c.Factory.KubernetesClientSet()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to create kubernetes client: %w", err)
	}

	return g, c, kc, nil
}
