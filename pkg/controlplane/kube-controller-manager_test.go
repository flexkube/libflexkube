package controlplane

import (
	"testing"

	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

const (
	nonEmptyString = "foo"
)

func TestKubeControllerManagerValidate(t *testing.T) {
	hostConfig := &host.Host{
		DirectConfig: &direct.Config{},
	}

	cases := map[string]struct {
		Config *KubeControllerManager
		Error  bool
	}{
		"require KubernetesCACertificate": {
			Config: &KubeControllerManager{
				KubernetesCAKey:          nonEmptyString,
				ServiceAccountPrivateKey: nonEmptyString,
				APIServer:                nonEmptyString,
				ClientCertificate:        nonEmptyString,
				ClientKey:                nonEmptyString,
				RootCACertificate:        nonEmptyString,
				Host:                     hostConfig,
			},
			Error: true,
		},
		"require KubernetesCAKey": {
			Config: &KubeControllerManager{
				KubernetesCACertificate:  nonEmptyString,
				ServiceAccountPrivateKey: nonEmptyString,
				APIServer:                nonEmptyString,
				ClientCertificate:        nonEmptyString,
				ClientKey:                nonEmptyString,
				RootCACertificate:        nonEmptyString,
				Host:                     hostConfig,
			},
			Error: true,
		},
		"require ServiceAccountPrivateKey": {
			Config: &KubeControllerManager{
				KubernetesCACertificate: nonEmptyString,
				KubernetesCAKey:         nonEmptyString,
				APIServer:               nonEmptyString,
				ClientCertificate:       nonEmptyString,
				ClientKey:               nonEmptyString,
				RootCACertificate:       nonEmptyString,
				Host:                    hostConfig,
			},
			Error: true,
		},
		"require APIServer": {
			Config: &KubeControllerManager{
				KubernetesCACertificate:  nonEmptyString,
				KubernetesCAKey:          nonEmptyString,
				ServiceAccountPrivateKey: nonEmptyString,
				ClientCertificate:        nonEmptyString,
				ClientKey:                nonEmptyString,
				RootCACertificate:        nonEmptyString,
				Host:                     hostConfig,
			},
			Error: true,
		},
		"require ClientCertificate": {
			Config: &KubeControllerManager{
				KubernetesCACertificate:  nonEmptyString,
				KubernetesCAKey:          nonEmptyString,
				ServiceAccountPrivateKey: nonEmptyString,
				APIServer:                nonEmptyString,
				ClientKey:                nonEmptyString,
				RootCACertificate:        nonEmptyString,
				Host:                     hostConfig,
			},
			Error: true,
		},
		"require ClientKey": {
			Config: &KubeControllerManager{
				KubernetesCACertificate:  nonEmptyString,
				KubernetesCAKey:          nonEmptyString,
				ServiceAccountPrivateKey: nonEmptyString,
				APIServer:                nonEmptyString,
				ClientCertificate:        nonEmptyString,
				RootCACertificate:        nonEmptyString,
				Host:                     hostConfig,
			},
			Error: true,
		},
		"require RootCACertificate": {
			Config: &KubeControllerManager{
				KubernetesCACertificate:  nonEmptyString,
				KubernetesCAKey:          nonEmptyString,
				ServiceAccountPrivateKey: nonEmptyString,
				APIServer:                nonEmptyString,
				ClientCertificate:        nonEmptyString,
				ClientKey:                nonEmptyString,
				Host:                     hostConfig,
			},
			Error: true,
		},
		"no host": {
			Config: &KubeControllerManager{
				KubernetesCACertificate:  nonEmptyString,
				KubernetesCAKey:          nonEmptyString,
				ServiceAccountPrivateKey: nonEmptyString,
				APIServer:                nonEmptyString,
				ClientCertificate:        nonEmptyString,
				ClientKey:                nonEmptyString,
				RootCACertificate:        nonEmptyString,
			},
			Error: true,
		},
		"bad host": {
			Config: &KubeControllerManager{
				KubernetesCACertificate:  nonEmptyString,
				KubernetesCAKey:          nonEmptyString,
				ServiceAccountPrivateKey: nonEmptyString,
				APIServer:                nonEmptyString,
				ClientCertificate:        nonEmptyString,
				ClientKey:                nonEmptyString,
				RootCACertificate:        nonEmptyString,
				Host:                     &host.Host{},
			},
			Error: true,
		},
		"valid": {
			Config: &KubeControllerManager{
				KubernetesCACertificate:  nonEmptyString,
				KubernetesCAKey:          nonEmptyString,
				ServiceAccountPrivateKey: nonEmptyString,
				APIServer:                nonEmptyString,
				ClientCertificate:        nonEmptyString,
				ClientKey:                nonEmptyString,
				RootCACertificate:        nonEmptyString,
				Host:                     hostConfig,
			},
			Error: false,
		},
	}

	for n, c := range cases {
		c := c

		t.Run(n, func(t *testing.T) {
			if err := c.Config.Validate(); !c.Error && err != nil {
				t.Errorf("Didn't expect error, got: %v", err)
			}
		})
	}
}

func TestKubeControllerManagerNewFillImage(t *testing.T) {
	kcm := &KubeControllerManager{
		KubernetesCACertificate:  nonEmptyString,
		KubernetesCAKey:          nonEmptyString,
		ServiceAccountPrivateKey: nonEmptyString,
		APIServer:                nonEmptyString,
		ClientCertificate:        nonEmptyString,
		ClientKey:                nonEmptyString,
		RootCACertificate:        nonEmptyString,
		Host: &host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	o, err := kcm.New()
	if err != nil {
		t.Fatalf("new should not return error, got: %v", err)
	}

	if o.image == "" {
		t.Fatalf("New() should set default image if it's not present")
	}
}

func TestKubeControllerManagerToHostConfiguredContainer(t *testing.T) {
	kcm := &KubeControllerManager{
		KubernetesCACertificate:  nonEmptyString,
		KubernetesCAKey:          nonEmptyString,
		ServiceAccountPrivateKey: nonEmptyString,
		APIServer:                nonEmptyString,
		ClientCertificate:        nonEmptyString,
		ClientKey:                nonEmptyString,
		RootCACertificate:        nonEmptyString,
		Host: &host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	o, err := kcm.New()
	if err != nil {
		t.Fatalf("new should not return error, got: %v", err)
	}

	// TODO grab an object and perform some validation on it?
	o.ToHostConfiguredContainer()
}
