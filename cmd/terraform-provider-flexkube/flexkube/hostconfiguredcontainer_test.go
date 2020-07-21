package flexkube

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

func TestHostConfiguredContainerMarshal(t *testing.T) { //nolint:funlen
	c := container.HostConfiguredContainer{
		Container: container.Container{
			Config: types.ContainerConfig{},
			Runtime: container.RuntimeConfig{
				Docker: docker.DefaultConfig(),
			},
		},
		ConfigFiles: map[string]string{
			"/foo": "bar",
		},
		Host: host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	var s []interface{}

	e := map[string]interface{}{
		"name": "foo",
		"config_files": map[string]interface{}{
			"/foo": "bar",
		},
		"host": []interface{}{
			map[string]interface{}{
				"direct": []interface{}{
					map[string]interface{}{},
				},
			},
		},
		"container": []interface{}{
			map[string]interface{}{
				"config": []interface{}{
					map[string]interface{}{
						"name":         "",
						"image":        "",
						"privileged":   false,
						"args":         s,
						"entrypoint":   s,
						"port":         []interface{}{},
						"mount":        []interface{}{},
						"network_mode": "",
						"pid_mode":     "",
						"ipc_mode":     "",
						"user":         "",
						"group":        "",
					},
				},
				"runtime": []interface{}{
					map[string]interface{}{
						"docker": []interface{}{
							map[string]interface{}{
								"host": "unix:///var/run/docker.sock",
							},
						},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(hostConfiguredContainerMarshal("foo", c, false), e); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func hostConfiguredContainerMarshaled() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"name": "foo",
			"config_files": map[string]interface{}{
				"/foo": "bar",
			},
			"host": []interface{}{
				map[string]interface{}{
					"direct": []interface{}{
						map[string]interface{}{},
					},
				},
			},
			"container": []interface{}{
				map[string]interface{}{
					"config": []interface{}{
						map[string]interface{}{
							"name":         "",
							"image":        "",
							"privileged":   false,
							"args":         []interface{}{},
							"entrypoint":   []interface{}{},
							"port":         []interface{}{},
							"mount":        []interface{}{},
							"network_mode": "",
							"pid_mode":     "",
							"ipc_mode":     "",
							"user":         "",
							"group":        "",
						},
					},
					"runtime": []interface{}{
						map[string]interface{}{
							"docker": []interface{}{
								map[string]interface{}{
									"host": "unix:///var/run/docker.sock",
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestHostConfiguredContainerUnmarshal(t *testing.T) {
	c := container.HostConfiguredContainer{
		Container: container.Container{
			Config: types.ContainerConfig{},
			Runtime: container.RuntimeConfig{
				Docker: docker.DefaultConfig(),
			},
		},
		ConfigFiles: map[string]string{
			"/foo": "bar",
		},
		Host: host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	e := hostConfiguredContainerMarshaled()

	n, d := hostConfiguredContainerUnmarshal(e[0])

	if diff := cmp.Diff(d, &c); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}

	if n != "foo" {
		t.Fatalf("Name should be %q, got: %q", "foo", n)
	}
}
