package kubelet

import (
	"testing"

	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

func TestToHostConfiguredContainer(t *testing.T) {
	kk := &Kubelet{
		BootstrapKubeconfig: "foo",
		Host: host.Host{
			DirectConfig: &direct.Config{},
		},
		Labels: map[string]string{
			"foo": "bar",
		},
		Taints: map[string]string{
			"foo": "bar",
		},
		PrivilegedLabels: map[string]string{
			"baz": "bar",
		},
		PrivilegedLabelsKubeconfig: "foo",
		ClusterDNSIPs:              []string{"10.0.0.1"},
	}

	k, err := kk.New()
	if err != nil {
		t.Fatalf("Creating new kubelet should succeed, got: %v", err)
	}

	hcc := k.ToHostConfiguredContainer()

	if _, err := hcc.New(); err != nil {
		t.Fatalf("should produce valid HostConfiguredContainer, got: %v", err)
	}
}
