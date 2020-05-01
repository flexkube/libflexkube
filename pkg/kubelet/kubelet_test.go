package kubelet

import (
	"reflect"
	"testing"

	"github.com/flexkube/libflexkube/internal/utiltest"
	containertypes "github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
	"github.com/flexkube/libflexkube/pkg/types"
)

func TestToHostConfiguredContainer(t *testing.T) {
	kk := &Kubelet{
		BootstrapKubeconfig:     "foo",
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
		PrivilegedLabelsKubeconfig: "foo",
		ClusterDNSIPs:              []string{"10.0.0.1"},
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
func TestKubeletValidate(t *testing.T) {
	k := &Kubelet{
		BootstrapKubeconfig:     "foo",
		Name:                    "foo",
		NetworkPlugin:           "cni",
		VolumePluginDir:         "/foo",
		KubernetesCACertificate: types.Certificate(utiltest.GenerateX509Certificate(t)),
		Host: host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	if err := k.Validate(); err != nil {
		t.Fatalf("validation of kubelet should pass, got: %v", err)
	}
}

func TestKubeletValidateRequireName(t *testing.T) {
	k := &Kubelet{
		BootstrapKubeconfig: "foo",
		NetworkPlugin:       "cni",
		VolumePluginDir:     "/foo",
		Host: host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	if err := k.Validate(); err == nil {
		t.Fatalf("validation of kubelet should fail when name is not set")
	}
}

func TestKubeletValidateEmptyCA(t *testing.T) {
	k := &Kubelet{
		BootstrapKubeconfig: "foo",
		NetworkPlugin:       "cni",
		VolumePluginDir:     "/foo",
		Host: host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	if err := k.Validate(); err == nil {
		t.Fatalf("validation of kubelet should fail when kubernetes CA certificate is not set")
	}
}

func TestKubeletValidateBadCA(t *testing.T) {
	k := &Kubelet{
		BootstrapKubeconfig:     "foo",
		NetworkPlugin:           "cni",
		VolumePluginDir:         "/foo",
		KubernetesCACertificate: "doh",
		Host: host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	if err := k.Validate(); err == nil {
		t.Fatalf("validation of kubelet should fail when kubernetes CA certificate is not valid")
	}
}

func TestKubeletIncludeExtraMounts(t *testing.T) {
	em := containertypes.Mount{
		Source: "/tmp/",
		Target: "/foo",
	}

	kk := &Kubelet{
		BootstrapKubeconfig:     "foo",
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
		ExtraMounts:                []containertypes.Mount{em},
		PrivilegedLabelsKubeconfig: "foo",
		ClusterDNSIPs:              []string{"10.0.0.1"},
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
