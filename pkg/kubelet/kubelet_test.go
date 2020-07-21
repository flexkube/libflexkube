package kubelet

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/flexkube/libflexkube/internal/utiltest"
	containertypes "github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
	"github.com/flexkube/libflexkube/pkg/pki"
	"github.com/flexkube/libflexkube/pkg/types"
)

func getClientConfig(t *testing.T) *client.Config {
	t.Parallel()

	p := &pki.PKI{
		Kubernetes: &pki.Kubernetes{},
	}

	if err := p.Generate(); err != nil {
		t.Fatalf("failed generating testing PKI: %v", err)
	}

	return &client.Config{
		Server:        "foo",
		CACertificate: p.Kubernetes.CA.X509Certificate,
		Token:         "foob",
	}
}

func TestToHostConfiguredContainer(t *testing.T) {
	cc := getClientConfig(t)

	kk := &Kubelet{
		BootstrapConfig:         cc,
		Name:                    "fooz",
		NetworkPlugin:           "cni",
		VolumePluginDir:         "/var/lib/kubelet/volumeplugins",
		KubernetesCACertificate: types.Certificate(utiltest.GenerateX509Certificate(t)),
		Host: host.Host{
			DirectConfig: &direct.Config{},
		},
		Labels: map[string]string{
			"do": "bar",
		},
		Taints: map[string]string{
			"noh": "bar",
		},
		PrivilegedLabels: map[string]string{
			"baz": "bar",
		},

		AdminConfig:   cc,
		ClusterDNSIPs: []string{"10.0.0.1"},
	}

	k, err := kk.New()
	if err != nil {
		t.Fatalf("Creating new kubelet should succeed, got: %v", err)
	}

	hcc, err := k.ToHostConfiguredContainer()
	if err != nil {
		t.Fatalf("Generating HostConfiguredContainer should work, got: %v", err)
	}

	if _, err := hcc.New(); err != nil {
		t.Fatalf("should produce valid HostConfiguredContainer, got: %v", err)
	}
}

// Validate() tests.
func TestKubeletValidate(t *testing.T) { //nolint:funlen
	t.Parallel()

	cases := []struct {
		MutationF func(k *Kubelet)
		TestF     func(t *testing.T, er error)
	}{
		{
			MutationF: func(k *Kubelet) {},
			TestF: func(t *testing.T, err error) {
				if err != nil {
					t.Fatalf("validation of kubelet should pass, got: %v", err)
				}
			},
		},
		{
			MutationF: func(k *Kubelet) { k.Name = "" },
			TestF: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("validation of kubelet should fail when name is not set")
				}
			},
		},
		{
			MutationF: func(k *Kubelet) {},
			TestF:     func(t *testing.T, err error) {},
		},
		{
			MutationF: func(k *Kubelet) { k.KubernetesCACertificate = "" },
			TestF: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("validation of kubelet should fail when kubernetes CA certificate is not set")
				}
			},
		},
		{
			MutationF: func(k *Kubelet) { k.BootstrapConfig.Server = "" },
			TestF: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("validation of kubelet should fail when bootstrap config is invalid")
				}
			},
		},
		{
			MutationF: func(k *Kubelet) { k.VolumePluginDir = "" },
			TestF: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("validation of kubelet should fail when volume plugin dir is empty")
				}
			},
		},
		{
			MutationF: func(k *Kubelet) {
				k.PrivilegedLabels = map[string]string{
					"foo": "bar",
				}
				k.AdminConfig = nil
			},
			TestF: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("validation of kubelet should fail when privileged labels are configured and admin config is not")
				}
			},
		},
		{
			MutationF: func(k *Kubelet) { k.AdminConfig = k.BootstrapConfig },
			TestF: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("validation of kubelet should fail when admin config is defined and there is no privileged labels")
				}
			},
		},
		{
			MutationF: func(k *Kubelet) {
				k.PrivilegedLabels = map[string]string{
					"foo": "bar",
				}
				k.AdminConfig = &client.Config{}
			},
			TestF: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("validation of kubelet should fail when admin config is wrong")
				}
			},
		},
		{
			MutationF: func(k *Kubelet) { k.PodCIDR = "foo" },
			TestF: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("validation of kubelet should fail when network plugin is 'cni' and pod CIDR is set")
				}
			},
		},
		{
			MutationF: func(k *Kubelet) { k.NetworkPlugin = KubenetNetworkPlugin },
			TestF: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("validation of kubelet should fail when network plugin is 'kubelet' and pod CIDR is empty")
				}
			},
		},
		{
			MutationF: func(k *Kubelet) { k.NetworkPlugin = "doh" },
			TestF: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("validation of kubelet should fail when network plugin is invalid")
				}
			},
		},
		{
			MutationF: func(k *Kubelet) { k.Host.DirectConfig = nil },
			TestF: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("validation of kubelet should fail when host is invalid")
				}
			},
		},
	}

	for i, tc := range cases {
		tc := tc

		t.Run(strconv.Itoa(i), func(t *testing.T) {
			cc := getClientConfig(t)

			k := &Kubelet{
				BootstrapConfig:         cc,
				Name:                    "foo",
				NetworkPlugin:           "cni",
				VolumePluginDir:         "/foo",
				KubernetesCACertificate: types.Certificate(utiltest.GenerateX509Certificate(t)),
				Host: host.Host{
					DirectConfig: &direct.Config{},
				},
			}

			tc.MutationF(k)

			tc.TestF(t, k.Validate())
		})
	}
}

func TestKubeletIncludeExtraMounts(t *testing.T) {
	em := containertypes.Mount{
		Source: "/tmp/",
		Target: "/foo",
	}

	cc := getClientConfig(t)

	kk := &Kubelet{
		BootstrapConfig:         cc,
		Name:                    "foo",
		NetworkPlugin:           "cni",
		VolumePluginDir:         "/var/lib/kubelet/volumeplugins",
		KubernetesCACertificate: types.Certificate(utiltest.GenerateX509Certificate(t)),
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
		ExtraMounts:   []containertypes.Mount{em},
		AdminConfig:   cc,
		ClusterDNSIPs: []string{"10.0.0.1"},
	}

	k, err := kk.New()
	if err != nil {
		t.Fatalf("Creating new kubelet should succeed, got: %v", err)
	}

	found := false

	for _, v := range k.(*kubelet).mounts() {
		if reflect.DeepEqual(v, em) {
			found = true
		}
	}

	if !found {
		t.Fatalf("extra mount should be included in generated mounts")
	}
}
