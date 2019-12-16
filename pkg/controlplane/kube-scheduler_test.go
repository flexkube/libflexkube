package controlplane

import (
	"testing"

	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

func TestKubeSchedulerToHostConfiguredContainer(t *testing.T) {
	ks := &KubeScheduler{
		KubernetesCACertificate: nonEmptyString,
		APIServer:               nonEmptyString,
		ClientCertificate:       nonEmptyString,
		ClientKey:               nonEmptyString,
		FrontProxyCACertificate: nonEmptyString,
		Host: &host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	o, err := ks.New()
	if err != nil {
		t.Fatalf("new should not return error, got: %v", err)
	}

	// TODO grab an object and perform some validation on it?
	o.ToHostConfiguredContainer()
}
