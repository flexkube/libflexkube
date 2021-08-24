package controlplane

import (
	"strings"
	"testing"

	"github.com/flexkube/libflexkube/internal/utiltest"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
	"github.com/flexkube/libflexkube/pkg/types"
)

const (
	// TLS port used for testing.
	securePort = 6443

	// Non empty string used for testing.
	nonEmptyString = "foo"
)

func TestKubeAPIServerToHostConfiguredContainer(t *testing.T) {
	t.Parallel()

	cert := types.Certificate(utiltest.GenerateX509Certificate(t))
	privateKey := types.PrivateKey(utiltest.GenerateRSAPrivateKey(t))

	kas := &KubeAPIServer{
		Common: &Common{
			KubernetesCACertificate: cert,
			FrontProxyCACertificate: cert,
		},
		APIServerCertificate:     cert,
		APIServerKey:             privateKey,
		ServiceAccountPrivateKey: nonEmptyString,
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
		Host: &host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	o, err := kas.New()
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

func validKubeAPIServer(t *testing.T) *KubeAPIServer {
	t.Helper()

	cert := types.Certificate(utiltest.GenerateX509Certificate(t))
	privateKey := types.PrivateKey(utiltest.GenerateRSAPrivateKey(t))

	hostConfig := &host.Host{
		DirectConfig: &direct.Config{},
	}

	common := &Common{
		KubernetesCACertificate: cert,
		FrontProxyCACertificate: cert,
	}

	return &KubeAPIServer{
		Common:                   common,
		APIServerCertificate:     cert,
		APIServerKey:             privateKey,
		ServiceAccountPrivateKey: nonEmptyString,
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
	}
}

// Validate() tests.
func TestKubeAPIServerValidate(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		MutateF func(*KubeAPIServer)
		Error   bool
	}{
		"require kubeletClientCertificate": {
			MutateF: func(k *KubeAPIServer) {
				k.KubeletClientCertificate = ""
			},
			Error: true,
		},
		"validate kubeletClientCertificate": {
			MutateF: func(k *KubeAPIServer) {
				k.KubeletClientCertificate = nonEmptyString
			},
			Error: true,
		},
		"require at least one etcd server": {
			MutateF: func(k *KubeAPIServer) {
				k.EtcdServers = []string{}
			},
			Error: true,
		},
		"validate host": {
			MutateF: func(k *KubeAPIServer) {
				k.Host = &host.Host{}
			},
			Error: true,
		},
		"valid": {
			MutateF: func(_ *KubeAPIServer) {},
			Error:   false,
		},
	}

	for n, c := range cases {
		c := c

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			config := validKubeAPIServer(t)
			c.MutateF(config)

			err := config.Validate()
			if !c.Error && err != nil {
				t.Errorf("didn't expect error, got: %v", err)
			}

			if c.Error && err == nil {
				t.Errorf("expected error")
			}
		})
	}
}

func TestKubeAPIServerConfigFiles(t *testing.T) {
	t.Parallel()

	cert := types.Certificate(utiltest.GenerateX509Certificate(t))
	privateKey := types.PrivateKey(utiltest.GenerateRSAPrivateKey(t))

	hostConfig := &host.Host{
		DirectConfig: &direct.Config{},
	}

	common := &Common{
		KubernetesCACertificate: cert,
		FrontProxyCACertificate: cert,
	}

	c := &KubeAPIServer{
		Common:                   common,
		APIServerCertificate:     cert,
		APIServerKey:             privateKey,
		ServiceAccountPrivateKey: nonEmptyString,
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
	}

	ki, err := c.New()
	if err != nil {
		t.Fatalf("kubeAPIServer object should be created, got: %v", err)
	}

	hcc, err := ki.ToHostConfiguredContainer()
	if err != nil {
		t.Fatalf("Converting kube-apiserver to host configured container: %v", err)
	}

	for k := range hcc.ConfigFiles {
		if !strings.Contains(k, hostConfigPath) {
			t.Fatalf("all config files paths should contain %s, got: %s", hostConfigPath, k)
		}
	}
}

// New() tests.
func TestKubeAPIServerNewEmptyHost(t *testing.T) {
	t.Parallel()

	c := &KubeAPIServer{}

	k, err := c.New()
	if err == nil {
		t.Errorf("New on empty config should return error")
	}

	if k != nil {
		t.Errorf("New should not return kube-apiserver object in case of error")
	}
}
