package controlplane

import (
	"testing"

	"github.com/flexkube/libflexkube/internal/utiltest"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
	"github.com/flexkube/libflexkube/pkg/types"
)

func validValidator(t *testing.T) validator {
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

	return validator{
		Common:     common,
		Kubeconfig: kubeconfig,
		Host:       hostConfig,
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	v := validValidator(t)

	if err := v.validate(true); err != nil {
		t.Fatalf("validating valid object should succeed, got: %v", err)
	}
}

func TestValidateMarshalFail(t *testing.T) {
	t.Parallel()

	v := validValidator(t)

	v.YAML = map[string]interface{}{
		"foo": make(chan int),
	}

	if err := v.validate(true); err == nil {
		t.Fatalf("validating unmarshalable struct should fail")
	}
}
