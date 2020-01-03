package controlplane

import (
	"testing"

	"github.com/flexkube/libflexkube/internal/utiltest"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
	"github.com/flexkube/libflexkube/pkg/types"
)

func TestKubeSchedulerToHostConfiguredContainer(t *testing.T) {
	pki := utiltest.GeneratePKI(t)

	ks := &KubeScheduler{
		Common: Common{
			FrontProxyCACertificate: types.Certificate(pki.Certificate),
		},
		Kubeconfig: client.Config{
			Server:            "localhost",
			CACertificate:     types.Certificate(pki.Certificate),
			ClientCertificate: types.Certificate(pki.Certificate),
			ClientKey:         types.PrivateKey(pki.PrivateKey),
		},
		Host: host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	o, err := ks.New()
	if err != nil {
		t.Fatalf("new should not return error, got: %v", err)
	}

	// TODO grab an object and perform some validation on it?
	hcc := o.ToHostConfiguredContainer()

	if hcc.Container.Config.Image == "" {
		t.Fatalf("New() should set default image if it's not present")
	}
}
