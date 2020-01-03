package controlplane

import (
	"testing"

	"github.com/flexkube/libflexkube/internal/utiltest"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
	"github.com/flexkube/libflexkube/pkg/types"
)

const (
	// securePort is a TLS port used for testing.
	securePort = 6443

	// nonEmptyString is a string used for testing.
	nonEmptyString = "foo"
)

func TestKubeAPIServerToHostConfiguredContainer(t *testing.T) {
	cert := types.Certificate(utiltest.GenerateX509Certificate(t))
	privateKey := types.PrivateKey(utiltest.GenerateRSAPrivateKey(t))

	kas := &KubeAPIServer{
		Common: Common{
			KubernetesCACertificate: cert,
			FrontProxyCACertificate: cert,
		},
		APIServerCertificate:     cert,
		APIServerKey:             privateKey,
		ServiceAccountPublicKey:  nonEmptyString,
		BindAddress:              nonEmptyString,
		AdvertiseAddress:         nonEmptyString,
		EtcdServers:              []string{nonEmptyString},
		ServiceCIDR:              nonEmptyString,
		SecurePort:               securePort,
		FrontProxyCertificate:    cert,
		FrontProxyKey:            privateKey,
		KubeletClientCertificate: cert,
		KubeletClientKey:         privateKey,
		EtcdCACertificate:        cert,
		EtcdClientCertificate:    cert,
		EtcdClientKey:            privateKey,
		Host: host.Host{
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

func TestKubeAPIServerValidate(t *testing.T) {
	cert := types.Certificate(utiltest.GenerateX509Certificate(t))
	privateKey := types.PrivateKey(utiltest.GenerateRSAPrivateKey(t))

	hostConfig := host.Host{
		DirectConfig: &direct.Config{},
	}

	common := Common{
		KubernetesCACertificate: cert,
		FrontProxyCACertificate: cert,
	}

	cases := map[string]struct {
		Config *KubeAPIServer
		Error  bool
	}{
		"require kubeletClientCertificate": {
			Config: &KubeAPIServer{
				Common:                  common,
				APIServerCertificate:    cert,
				APIServerKey:            privateKey,
				ServiceAccountPublicKey: nonEmptyString,
				BindAddress:             nonEmptyString,
				AdvertiseAddress:        nonEmptyString,
				EtcdServers:             []string{nonEmptyString},
				ServiceCIDR:             nonEmptyString,
				SecurePort:              securePort,
				FrontProxyCertificate:   cert,
				FrontProxyKey:           privateKey,
				KubeletClientKey:        privateKey,
				EtcdCACertificate:       cert,
				EtcdClientCertificate:   cert,
				EtcdClientKey:           privateKey,
				Host:                    hostConfig,
			},
			Error: true,
		},
		"validate kubeletClientCertificate": {
			Config: &KubeAPIServer{
				Common:                   common,
				APIServerCertificate:     cert,
				APIServerKey:             privateKey,
				ServiceAccountPublicKey:  nonEmptyString,
				BindAddress:              nonEmptyString,
				AdvertiseAddress:         nonEmptyString,
				EtcdServers:              []string{nonEmptyString},
				ServiceCIDR:              nonEmptyString,
				SecurePort:               securePort,
				FrontProxyCertificate:    cert,
				FrontProxyKey:            privateKey,
				KubeletClientKey:         privateKey,
				EtcdCACertificate:        cert,
				EtcdClientCertificate:    cert,
				EtcdClientKey:            privateKey,
				Host:                     hostConfig,
				KubeletClientCertificate: nonEmptyString,
			},
			Error: true,
		},
		"valid": {
			Config: &KubeAPIServer{
				Common:                   common,
				APIServerCertificate:     cert,
				APIServerKey:             privateKey,
				ServiceAccountPublicKey:  nonEmptyString,
				BindAddress:              nonEmptyString,
				AdvertiseAddress:         nonEmptyString,
				EtcdServers:              []string{nonEmptyString},
				ServiceCIDR:              nonEmptyString,
				SecurePort:               securePort,
				FrontProxyCertificate:    cert,
				FrontProxyKey:            privateKey,
				KubeletClientKey:         privateKey,
				EtcdCACertificate:        cert,
				EtcdClientCertificate:    cert,
				EtcdClientKey:            privateKey,
				Host:                     hostConfig,
				KubeletClientCertificate: cert,
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
