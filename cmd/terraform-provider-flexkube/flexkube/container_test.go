package flexkube

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
)

func TestContainerMarshal(t *testing.T) {
	c := container.Container{
		Config: types.ContainerConfig{},
		Runtime: container.RuntimeConfig{
			Docker: docker.DefaultConfig(),
		},
	}

	var s []interface{}

	e := []interface{}{
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
	}

	if diff := cmp.Diff(containerMarshal(c), e); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestContainerMarshalWithStatus(t *testing.T) {
	c := container.Container{
		Config: types.ContainerConfig{},
		Runtime: container.RuntimeConfig{
			Docker: docker.DefaultConfig(),
		},
		Status: &types.ContainerStatus{
			ID:     "foo",
			Status: "running",
		},
	}

	var s []interface{}

	e := []interface{}{
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
			"status": []interface{}{
				map[string]interface{}{
					"id":     "foo",
					"status": "running",
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
	}

	if diff := cmp.Diff(containerMarshal(c), e); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestContainerUnmarshal(t *testing.T) {
	c := container.Container{
		Config: types.ContainerConfig{},
		Runtime: container.RuntimeConfig{
			Docker: docker.DefaultConfig(),
		},
		Status: &types.ContainerStatus{
			ID:     "foo",
			Status: "running",
		},
	}

	e := []interface{}{
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
			"status": []interface{}{
				map[string]interface{}{
					"id":     "foo",
					"status": "running",
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
	}

	if diff := cmp.Diff(containerUnmarshal(e[0]), c); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}
