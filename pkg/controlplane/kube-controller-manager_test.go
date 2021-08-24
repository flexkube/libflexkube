package controlplane

import (
	"testing"

	"github.com/flexkube/libflexkube/internal/utiltest"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
	"github.com/flexkube/libflexkube/pkg/types"
)

//nolint:funlen // Just a lot of test cases.
func TestKubeControllerManagerValidate(t *testing.T) {
	t.Parallel()

	hostConfig := &host.Host{
		DirectConfig: &direct.Config{},
	}

	pki := utiltest.GeneratePKI(t)

	kubeconfig := client.Config{
		Server:            "localhost",
		CACertificate:     types.Certificate(pki.Certificate),
		ClientCertificate: types.Certificate(pki.Certificate),
		ClientKey:         types.PrivateKey(pki.PrivateKey),
	}

	common := &Common{
		KubernetesCACertificate: types.Certificate(pki.Certificate),
		FrontProxyCACertificate: types.Certificate(pki.Certificate),
	}

	cases := map[string]struct {
		Config *KubeControllerManager
		Error  bool
	}{
		"require Kubeconfig": {
			Config: &KubeControllerManager{
				KubernetesCAKey:          types.PrivateKey(pki.PrivateKey),
				ServiceAccountPrivateKey: types.PrivateKey(pki.PrivateKey),
				RootCACertificate:        types.Certificate(pki.Certificate),
				Host:                     hostConfig,
				Common:                   common,
			},
			Error: true,
		},
		"require KubernetesCAKey": {
			Config: &KubeControllerManager{
				ServiceAccountPrivateKey: types.PrivateKey(pki.PrivateKey),
				RootCACertificate:        types.Certificate(pki.Certificate),
				Host:                     hostConfig,
				Kubeconfig:               kubeconfig,
				Common:                   common,
			},
			Error: true,
		},
		"require ServiceAccountPrivateKey": {
			Config: &KubeControllerManager{
				KubernetesCAKey:   types.PrivateKey(pki.PrivateKey),
				RootCACertificate: types.Certificate(pki.Certificate),
				Host:              hostConfig,
				Kubeconfig:        kubeconfig,
				Common:            common,
			},
			Error: true,
		},
		"require RootCACertificate": {
			Config: &KubeControllerManager{
				KubernetesCAKey:          types.PrivateKey(pki.PrivateKey),
				ServiceAccountPrivateKey: types.PrivateKey(pki.PrivateKey),
				Host:                     hostConfig,
				Kubeconfig:               kubeconfig,
				Common:                   common,
			},
			Error: true,
		},
		"no host": {
			Config: &KubeControllerManager{
				KubernetesCAKey:          types.PrivateKey(pki.PrivateKey),
				ServiceAccountPrivateKey: types.PrivateKey(pki.PrivateKey),
				RootCACertificate:        types.Certificate(pki.Certificate),
				Kubeconfig:               kubeconfig,
				Common:                   common,
			},
			Error: true,
		},
		"bad host": {
			Config: &KubeControllerManager{
				KubernetesCAKey:          types.PrivateKey(pki.PrivateKey),
				ServiceAccountPrivateKey: types.PrivateKey(pki.PrivateKey),
				RootCACertificate:        types.Certificate(pki.Certificate),
				Kubeconfig:               kubeconfig,
				Host:                     &host.Host{},
			},
			Error: true,
		},
		"valid": {
			Config: &KubeControllerManager{
				KubernetesCAKey:          types.PrivateKey(pki.PrivateKey),
				ServiceAccountPrivateKey: types.PrivateKey(pki.PrivateKey),
				RootCACertificate:        types.Certificate(pki.Certificate),
				Host:                     hostConfig,
				Kubeconfig:               kubeconfig,
				Common:                   common,
			},
			Error: false,
		},
	}

	for n, c := range cases {
		c := c

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			err := c.Config.Validate()
			if !c.Error && err != nil {
				t.Errorf("Didn't expect error, got: %v", err)
			}

			if c.Error && err == nil {
				t.Errorf("Expected error")
			}
		})
	}
}

func TestKubeControllerManagerToHostConfiguredContainer(t *testing.T) {
	t.Parallel()

	pki := utiltest.GeneratePKI(t)

	kcm := &KubeControllerManager{
		KubernetesCAKey:          types.PrivateKey(pki.PrivateKey),
		ServiceAccountPrivateKey: types.PrivateKey(pki.PrivateKey),
		RootCACertificate:        types.Certificate(pki.Certificate),
		Host: &host.Host{
			DirectConfig: &direct.Config{},
		},
		Kubeconfig: client.Config{
			Server:            "localhost",
			CACertificate:     types.Certificate(pki.Certificate),
			ClientCertificate: types.Certificate(pki.Certificate),
			ClientKey:         types.PrivateKey(pki.PrivateKey),
		},
	}

	o, err := kcm.New()
	if err != nil {
		t.Fatalf("New should not return error, got: %v", err)
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
func TestKubeControllerManagerNewEmptyHost(t *testing.T) {
	t.Parallel()

	ks := &KubeControllerManager{}

	k, err := ks.New()
	if err == nil {
		t.Errorf("Attempting to create kube-scheduler from empty config should fail")
	}

	if k != nil {
		t.Fatalf("Failed attempt of creating kube-scheduler should not return kube-scheduler object")
	}
}
