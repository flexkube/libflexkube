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
		Common: &Common{
			FrontProxyCACertificate: types.Certificate(pki.Certificate),
		},
		Kubeconfig: client.Config{
			Server:            "localhost",
			CACertificate:     types.Certificate(pki.Certificate),
			ClientCertificate: types.Certificate(pki.Certificate),
			ClientKey:         types.PrivateKey(pki.PrivateKey),
		},
		Host: &host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	o, err := ks.New()
	if err != nil {
		t.Fatalf("new should not return error, got: %v", err)
	}

	hcc, err := o.ToHostConfiguredContainer()
	if err != nil {
		t.Fatalf("Generating HostConfiguredContainer should work, got: %v", err)
	}

	if _, err := hcc.New(); err != nil {
		t.Fatalf("ToHostConfiguredContainer() should generate valid HostConfiguredContainer, got: %v", err)
	}

	if hcc.Container.Config.Image == "" {
		t.Fatalf("New() should set default image if it's not present")
	}
}

// New() tests.
func TestKubeSchedulerNewEmptyHost(t *testing.T) {
	ks := &KubeScheduler{}

	k, err := ks.New()
	if err == nil {
		t.Errorf("attempting to create kube-scheduler from empty config should fail")
	}

	if k != nil {
		t.Fatalf("failed attempt of creating kube-scheduler should not return kube-scheduler object")
	}
}

// Validate() tests.
func TestKubeSchedulerValidate(t *testing.T) { //nolint:funlen
	pki := utiltest.GeneratePKI(t)

	hostConfig := &host.Host{
		DirectConfig: &direct.Config{},
	}

	common := &Common{
		KubernetesCACertificate: types.Certificate(pki.Certificate),
		FrontProxyCACertificate: types.Certificate(pki.Certificate),
	}

	kubeconfig := client.Config{
		Server:            "localhost",
		CACertificate:     types.Certificate(pki.Certificate),
		ClientCertificate: types.Certificate(pki.Certificate),
		ClientKey:         types.PrivateKey(pki.PrivateKey),
	}

	cases := map[string]struct {
		Config *KubeScheduler
		Error  bool
	}{
		"require common certificates": {
			Config: &KubeScheduler{
				Host:       hostConfig,
				Kubeconfig: kubeconfig,
			},
			Error: true,
		},
		"validate kubeletClientCertificate": {
			Config: &KubeScheduler{
				Common: common,
				Host:   hostConfig,
			},
			Error: true,
		},
		"validate host": {
			Config: &KubeScheduler{
				Common:     common,
				Kubeconfig: kubeconfig,
				Host:       &host.Host{},
			},
			Error: true,
		},
		"valid": {
			Config: &KubeScheduler{
				Common:     common,
				Kubeconfig: kubeconfig,
				Host:       hostConfig,
			},
			Error: false,
		},
	}

	for n, c := range cases {
		c := c

		t.Run(n, func(t *testing.T) {
			err := c.Config.Validate()
			if !c.Error && err != nil {
				t.Errorf("didn't expect error, got: %v", err)
			}

			if c.Error && err == nil {
				t.Errorf("expected error")
			}
		})
	}
}
