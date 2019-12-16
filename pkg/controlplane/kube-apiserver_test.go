package controlplane

import (
	"testing"

	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

func TestKubeAPIServerToHostConfiguredContainer(t *testing.T) {
	kas := &KubeAPIServer{
		KubernetesCACertificate:  nonEmptyString,
		APIServerCertificate:     nonEmptyString,
		APIServerKey:             nonEmptyString,
		ServiceAccountPublicKey:  nonEmptyString,
		BindAddress:              nonEmptyString,
		AdvertiseAddress:         nonEmptyString,
		EtcdServers:              []string{nonEmptyString},
		ServiceCIDR:              nonEmptyString,
		SecurePort:               6443,
		FrontProxyCACertificate:  nonEmptyString,
		FrontProxyCertificate:    nonEmptyString,
		FrontProxyKey:            nonEmptyString,
		KubeletClientCertificate: nonEmptyString,
		KubeletClientKey:         nonEmptyString,

		Host: &host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	o, err := kas.New()
	if err != nil {
		t.Fatalf("new should not return error, got: %v", err)
	}

	// TODO grab an object and perform some validation on it?
	o.ToHostConfiguredContainer()
}
