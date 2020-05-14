package flexkube

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/kubelet"
	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
)

func TestKubeletUnmarshal(t *testing.T) {
	c := []kubelet.Kubelet{
		{
			Address:                 "h",
			Image:                   "a",
			KubernetesCACertificate: "b",
			ClusterDNSIPs:           []string{},
			Name:                    "g",
			Taints:                  nil,
			Labels:                  nil,
			PrivilegedLabels:        nil,
			CgroupDriver:            "c",
			NetworkPlugin:           "d",
			SystemReserved:          nil,
			KubeReserved:            nil,
			HairpinMode:             "e",
			VolumePluginDir:         "f",
			BootstrapConfig: &client.Config{
				Server: "foo",
			},
			AdminConfig: &client.Config{
				Server: "bar",
			},
		},
	}

	e := []interface{}{
		map[string]interface{}{
			"image":                     "a",
			"kubernetes_ca_certificate": "b",
			"cgroup_driver":             "c",
			"network_plugin":            "d",
			"hairpin_mode":              "e",
			"volume_plugin_dir":         "f",
			"name":                      "g",
			"address":                   "h",
			"bootstrap_config": []interface{}{
				map[string]interface{}{
					"server": "foo",
				},
			},
			"admin_config": []interface{}{
				map[string]interface{}{
					"server": "bar",
				},
			},
		},
	}

	if diff := cmp.Diff(kubeletsUnmarshal(e), c); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}
