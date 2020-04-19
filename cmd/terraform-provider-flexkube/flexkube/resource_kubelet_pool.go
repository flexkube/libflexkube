package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/kubelet"
	"github.com/flexkube/libflexkube/pkg/types"
)

func resourceKubeletPool() *schema.Resource {
	return &schema.Resource{
		Create:        resourceCreate(kubeletPoolUnmarshal),
		Read:          resourceRead(kubeletPoolUnmarshal),
		Delete:        resourceDelete(kubeletPoolUnmarshal, "kubelet"),
		Update:        resourceCreate(kubeletPoolUnmarshal),
		CustomizeDiff: resourceDiff(kubeletPoolUnmarshal),
		Schema: withCommonFields(map[string]*schema.Schema{
			"image":                     optionalString(false),
			"ssh":                       sshSchema(false),
			"bootstrap_kubeconfig":      optionalString(false),
			"kubernetes_ca_certificate": optionalString(false),
			"cluster_dns_ips":           optionalStringList(false),
			"taints": optionalMapPrimitive(false, func(computed bool) *schema.Schema {
				return &schema.Schema{
					Type: schema.TypeString,
				}
			}),
			"labels": optionalMapPrimitive(false, func(computed bool) *schema.Schema {
				return &schema.Schema{
					Type: schema.TypeString,
				}
			}),
			"privileged_labels": optionalMapPrimitive(false, func(computed bool) *schema.Schema {
				return &schema.Schema{
					Type: schema.TypeString,
				}
			}),
			"privileged_labels_kubeconfig": sensitiveString(false),
			"cgroup_driver":                optionalString(false),
			"network_plugin":               optionalString(false),
			"system_reserved": optionalMapPrimitive(false, func(computed bool) *schema.Schema {
				return &schema.Schema{
					Type: schema.TypeString,
				}
			}),
			"kube_reserved": optionalMapPrimitive(false, func(computed bool) *schema.Schema {
				return &schema.Schema{
					Type: schema.TypeString,
				}
			}),
			"hairpin_mode":      optionalString(false),
			"volume_plugin_dir": optionalString(false),
			"kubelet":           kubeletSchema(),
			"extra_mount":       mountsSchema(false),
		}),
	}
}

func kubeletPoolUnmarshal(d getter, includeState bool) types.ResourceConfig {
	c := &kubelet.Pool{
		Image:                      d.Get("image").(string),
		BootstrapKubeconfig:        d.Get("bootstrap_kubeconfig").(string),
		KubernetesCACertificate:    types.Certificate(d.Get("kubernetes_ca_certificate").(string)),
		PrivilegedLabelsKubeconfig: d.Get("privileged_labels_kubeconfig").(string),
		CgroupDriver:               d.Get("cgroup_driver").(string),
		NetworkPlugin:              d.Get("network_plugin").(string),
		HairpinMode:                d.Get("hairpin_mode").(string),
		VolumePluginDir:            d.Get("volume_plugin_dir").(string),
		Kubelets:                   kubeletsUnmarshal(d.Get("kubelet")),
		ClusterDNSIPs:              stringListUnmarshal(d.Get("cluster_dns_ips")),
		Taints:                     stringMapUnmarshal(d.Get("taints")),
		Labels:                     stringMapUnmarshal(d.Get("labels")),
		PrivilegedLabels:           stringMapUnmarshal(d.Get("privileged_labels")),
		SystemReserved:             stringMapUnmarshal(d.Get("system_reserved")),
		KubeReserved:               stringMapUnmarshal(d.Get("kube_reserved")),
		ExtraMounts:                mountsUnmarshal(d.Get("extra_mount")),
	}

	if s := getState(d); includeState && s != nil {
		c.State = *s
	}

	if d, ok := d.GetOk("ssh"); ok && len(d.([]interface{})) == 1 {
		c.SSH = sshUnmarshal(d.([]interface{})[0])
	}

	return c
}
