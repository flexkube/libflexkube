package flexkube

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/container/types"
)

const examplePort = 8080

func TestContainerConfigMarshal(t *testing.T) {
	cc := types.ContainerConfig{
		Name:       "foo",
		Image:      "bar",
		Privileged: true,
		Args:       []string{"foo"},
		Entrypoint: []string{"bar"},
		Ports: []types.PortMap{
			{
				IP:       "127.0.0.1",
				Port:     examplePort,
				Protocol: "tcp",
			},
		},
		Mounts: []types.Mount{
			{
				Source:      "/foo",
				Target:      "/bar",
				Propagation: "bidirectional",
			},
		},
		NetworkMode: "host",
		PidMode:     "host",
		IpcMode:     "host",
		User:        "1000",
		Group:       "2000",
	}

	e := []interface{}{
		map[string]interface{}{
			"name":       "foo",
			"image":      "bar",
			"privileged": true,
			"args":       []interface{}{"foo"},
			"entrypoint": []interface{}{"bar"},
			"port": []interface{}{
				map[string]interface{}{
					"ip":       "127.0.0.1",
					"port":     examplePort,
					"protocol": "tcp",
				},
			},
			"mount": []interface{}{
				map[string]interface{}{
					"source":      "/foo",
					"target":      "/bar",
					"propagation": "bidirectional",
				},
			},
			"network_mode": "host",
			"pid_mode":     "host",
			"ipc_mode":     "host",
			"user":         "1000",
			"group":        "2000",
		},
	}

	if diff := cmp.Diff(containerConfigMarshal(cc), e); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestContainerConfigUnmarshal(t *testing.T) {
	cc := types.ContainerConfig{
		Name:       "foo",
		Image:      "bar",
		Privileged: true,
		Args:       []string{"foo"},
		Entrypoint: []string{"bar"},
		Ports: []types.PortMap{
			{
				IP:       "127.0.0.1",
				Port:     examplePort,
				Protocol: "tcp",
			},
		},
		Mounts: []types.Mount{
			{
				Source:      "/foo",
				Target:      "/bar",
				Propagation: "bidirectional",
			},
		},
		NetworkMode: "host",
		PidMode:     "host",
		IpcMode:     "host",
		User:        "1000",
		Group:       "2000",
	}

	e := []interface{}{
		map[string]interface{}{
			"name":       "foo",
			"image":      "bar",
			"privileged": true,
			"args":       []interface{}{"foo"},
			"entrypoint": []interface{}{"bar"},
			"port": []interface{}{
				map[string]interface{}{
					"ip":       "127.0.0.1",
					"port":     examplePort,
					"protocol": "tcp",
				},
			},
			"mount": []interface{}{
				map[string]interface{}{
					"source":      "/foo",
					"target":      "/bar",
					"propagation": "bidirectional",
				},
			},
			"network_mode": "host",
			"pid_mode":     "host",
			"ipc_mode":     "host",
			"user":         "1000",
			"group":        "2000",
		},
	}

	if diff := cmp.Diff(containerConfigUnmarshal(e[0]), cc); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}
