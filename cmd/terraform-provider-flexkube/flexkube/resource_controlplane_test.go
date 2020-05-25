package flexkube

import (
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/flexkube/libflexkube/internal/utiltest"
	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

func TestControlplanePlanOnly(t *testing.T) {
	t.Parallel()

	config := `
locals {
  controller_ips = ["1.1.1.1"]
  controller_names = ["controller01"]
  first_controller_ip = local.controller_ips[0]
  api_port = 6443
  bootstrap_api_bind = "0.0.0.0"
}

resource "flexkube_pki" "pki" {
  certificate {
    organization = "example"
  }

  etcd {
    peers   = zipmap(local.controller_names, local.controller_ips)
    servers = zipmap(local.controller_names, local.controller_ips)

    client_cns = [
      "root",
      "kube-apiserver",
      "prometheus",
    ]
  }

  kubernetes {
    kube_api_server {
      external_names = ["kube-apiserver.example.com"]
      server_ips     = concat(local.controller_ips, ["127.0.1.1", "11.0.0.1"])
    }
  }
}

resource "flexkube_controlplane" "bootstrap" {
  pki_yaml = flexkube_pki.pki.state_yaml

  kube_apiserver {
    service_cidr               = "11.0.0.0/24"
    etcd_servers               = formatlist("https://%s:2379", local.controller_ips)
    bind_address               = local.bootstrap_api_bind
    advertise_address          = local.first_controller_ip
    secure_port                = local.api_port
  }

  kube_controller_manager {
    flex_volume_plugin_dir      = "/var/lib/kubelet/volumeplugins"
  }

  api_server_address = local.first_controller_ip
  api_server_port    = local.api_port

  ssh {
    address            = "127.0.0.1"
    port               = 12345
    password           = "bar"
    connection_timeout = "1s"
    retry_interval     = "1s"
    retry_timeout      = "1s"
  }
}
`

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"flexkube": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestControlplaneCreateRuntimeError(t *testing.T) {
	t.Parallel()

	config := `
locals {
  controller_ips = ["1.1.1.1"]
  controller_names = ["controller01"]
  first_controller_ip = local.controller_ips[0]
  api_port = 6443
  bootstrap_api_bind = "0.0.0.0"
}

resource "flexkube_pki" "pki" {
  certificate {
    organization = "example"
  }

  etcd {
    peers   = zipmap(local.controller_names, local.controller_ips)
    servers = zipmap(local.controller_names, local.controller_ips)

    client_cns = [
      "root",
      "kube-apiserver",
      "prometheus",
    ]
  }

  kubernetes {
    kube_api_server {
      external_names = ["kube-apiserver.example.com"]
      server_ips     = concat(local.controller_ips, ["127.0.1.1", "11.0.0.1"])
    }
  }
}

resource "flexkube_controlplane" "bootstrap" {
  pki_yaml = flexkube_pki.pki.state_yaml

  kube_apiserver {
    service_cidr               = "11.0.0.0/24"
    etcd_servers               = formatlist("https://%s:2379", local.controller_ips)
    bind_address               = local.bootstrap_api_bind
    advertise_address          = local.first_controller_ip
    secure_port                = local.api_port

    host {
      ssh {
        address            = "127.0.0.1"
        port               = 12345
        password           = "bar"
        connection_timeout = "1s"
        retry_interval     = "1s"
        retry_timeout      = "1s"
      }
    }

    common {
      image = "bar"
    }
  }

  kube_controller_manager {
    flex_volume_plugin_dir      = "/var/lib/kubelet/volumeplugins"

    host {
      ssh {
        address            = "127.0.0.1"
        port               = 12345
        password           = "bar"
        connection_timeout = "1s"
        retry_interval     = "1s"
        retry_timeout      = "1s"
      }
    }

    common {
      image = "baz"
    }

  }

  kube_scheduler {
    host {
      ssh {
        address            = "127.0.0.1"
        port               = 12345
        password           = "bar"
        connection_timeout = "1s"
        retry_interval     = "1s"
        retry_timeout      = "1s"
      }
    }

    common {
      image = "doh"
    }
  }

  api_server_address = local.first_controller_ip
  api_server_port    = local.api_port

  ssh {
    address            = "127.0.0.1"
    port               = 12345
    password           = "bar"
    connection_timeout = "1s"
    retry_interval     = "1s"
    retry_timeout      = "1s"
  }
}
`

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"flexkube": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`connection refused`),
			},
		},
	})
}

func TestControlplaneValidateFail(t *testing.T) {
	t.Parallel()

	config := `
resource "flexkube_controlplane" "bootstrap" {
  common {
    kubernetes_ca_certificate  = ""
    front_proxy_ca_certificate = ""
  }

  kube_apiserver {
    api_server_certificate     = ""
    api_server_key             = ""
    front_proxy_certificate    = ""
    front_proxy_key            = ""
    kubelet_client_certificate = ""
    kubelet_client_key         = ""
    service_account_public_key = ""
    etcd_ca_certificate        = ""
    etcd_client_certificate    = ""
    etcd_client_key            = ""
    service_cidr               = "11.0.0.0/24"
    etcd_servers               = []
    bind_address               = ""
    advertise_address          = ""
    secure_port                = 6443
  }

  kube_controller_manager {
    flex_volume_plugin_dir      = ""
    kubernetes_ca_key           = ""
    service_account_private_key = ""
    root_ca_certificate         = ""

    kubeconfig {
      client_certificate = ""
      client_key         = ""
    }
  }

  kube_scheduler {
    kubeconfig {
      client_certificate = ""
      client_key         = ""
    }
  }

  api_server_address = ""
  api_server_port    = 0

  ssh {
    user        = "core"
    address     = ""
    port        = 22
    private_key = ""
  }
}
`

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"flexkube": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`failed to decode PEM format`),
			},
		},
	})
}

func TestControlplaneDecodeEmptyBlocks(t *testing.T) {
	t.Parallel()

	config := `
resource "flexkube_controlplane" "bootstrap" {
  common {}

  kube_apiserver {}

  kube_controller_manager {}

  kube_scheduler {}

  ssh {}
}
`

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"flexkube": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`failed initializing configuration`),
			},
		},
	})
}

func TestControlplaneDecodeEmptyConfig(t *testing.T) {
	t.Parallel()

	config := `
resource "flexkube_controlplane" "bootstrap" {}
`

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"flexkube": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`required field is not set`),
			},
		},
	})
}

func TestControlplaneDecodeEmptyKubeconfig(t *testing.T) {
	t.Parallel()

	config := `
resource "flexkube_controlplane" "bootstrap" {
  kube_controller_manager {
    kubeconfig {}
    common {}
  }
}
`

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"flexkube": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`required field is not set`),
			},
		},
	})
}

func TestControlplaneDestroy(t *testing.T) {
	t.Parallel()

	pki := utiltest.GeneratePKI(t)

	// Prepare some fake state.
	cs := container.ContainersState{
		"foo": &container.HostConfiguredContainer{
			Host: host.Host{
				DirectConfig: &direct.Config{},
			},
			Container: container.Container{
				Runtime: container.RuntimeConfig{
					Docker: &docker.Config{
						Host: "unix:///nonexistent",
					},
				},
				Config: types.ContainerConfig{
					Name:  "foo",
					Image: "busybox:latest",
				},
				Status: &types.ContainerStatus{
					ID:     "foo",
					Status: "running",
				},
			},
		},
	}

	s := map[string]interface{}{
		stateSensitiveSchemaKey: containersStateMarshal(cs, false),
		"common": []interface{}{
			map[string]interface{}{
				"kubernetes_ca_certificate":  pki.Certificate,
				"front_proxy_ca_certificate": pki.Certificate,
			},
		},
		"kube_apiserver": []interface{}{
			map[string]interface{}{
				"api_server_certificate":     pki.Certificate,
				"api_server_key":             pki.PrivateKey,
				"front_proxy_certificate":    pki.Certificate,
				"front_proxy_key":            pki.PrivateKey,
				"kubelet_client_certificate": pki.Certificate,
				"kubelet_client_key":         pki.PrivateKey,
				"service_account_public_key": "foo",
				"service_cidr":               "11.0.0.0/24",
				"etcd_ca_certificate":        pki.Certificate,
				"etcd_client_certificate":    pki.Certificate,
				"etcd_client_key":            pki.PrivateKey,
				"etcd_servers": []interface{}{
					"foo",
				},
				"bind_address":      "0.0.0.0",
				"advertise_address": "1.1.1.1",
			},
		},
		"kube_controller_manager": []interface{}{
			map[string]interface{}{
				"kubernetes_ca_key":           pki.PrivateKey,
				"service_account_private_key": pki.PrivateKey,
				"kubeconfig": []interface{}{
					map[string]interface{}{
						"client_certificate": pki.Certificate,
						"client_key":         pki.PrivateKey,
					},
				},
				"root_ca_certificate": pki.Certificate,
			},
		},
		"api_server_address": "1.1.1.1",
		"api_server_port":    1,
		"kube_scheduler": []interface{}{
			map[string]interface{}{
				"kubeconfig": []interface{}{
					map[string]interface{}{
						"client_certificate": pki.Certificate,
						"client_key":         pki.PrivateKey,
					},
				},
			},
		},
	}

	r := resourceControlplane()
	d := schema.TestResourceDataRaw(t, r.Schema, s)

	// Mark newly created object as created, so it's state is persisted.
	d.SetId("foo")

	// Create new ResourceData from the state, so it's persisted and there is no diff included.
	dn := r.Data(d.State())

	err := controlplaneDestroy(dn, nil)
	if err == nil {
		t.Fatalf("destroying with unreachable container runtime should fail")
	}

	if !strings.Contains(err.Error(), "Is the docker daemon running") {
		t.Fatalf("destroying should fail for unreachable runtime, got: %v", err)
	}
}

func TestControlplaneDestroyValidateConfiguration(t *testing.T) {
	t.Parallel()

	// Prepare some fake state.
	cs := container.ContainersState{
		"foo": &container.HostConfiguredContainer{
			Host: host.Host{
				DirectConfig: &direct.Config{},
			},
			Container: container.Container{
				Runtime: container.RuntimeConfig{
					Docker: &docker.Config{
						Host: "unix:///nonexistent",
					},
				},
				Config: types.ContainerConfig{
					Name:  "foo",
					Image: "busybox:latest",
				},
				Status: &types.ContainerStatus{
					ID:     "foo",
					Status: "running",
				},
			},
		},
	}

	s := map[string]interface{}{
		stateSensitiveSchemaKey: containersStateMarshal(cs, false),
	}

	r := resourceControlplane()
	d := schema.TestResourceDataRaw(t, r.Schema, s)

	// Mark newly created object as created, so it's state is persisted.
	d.SetId("foo")

	// Create new ResourceData from the state, so it's persisted and there is no diff included.
	dn := r.Data(d.State())

	err := controlplaneDestroy(dn, nil)
	if err == nil {
		t.Fatalf("destroying with unreachable container runtime should fail")
	}

	if !strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
		t.Fatalf("destroying should fail for unreachable runtime, got: %v", err)
	}
}
