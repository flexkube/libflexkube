// +build integration

package container

import (
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

const (
	// containerRunningDelay is how long we wait for container to start and report as running by Docker.
	containerRunningDelay = 3 * time.Second
)

// Create() tests.
//
//nolint:funlen
func TestHostConfiguredContainerDeployConfigFile(t *testing.T) {
	t.Parallel()

	p := "/tmp/foo"
	f := path.Join(p, randomContainerName())

	h := &HostConfiguredContainer{
		Host: host.Host{
			DirectConfig: &direct.Config{},
		},
		Container: Container{
			Runtime: RuntimeConfig{
				Docker: &docker.Config{},
			},
			Config: types.ContainerConfig{
				Name:  "foo",
				Image: "nginx",
				Mounts: []types.Mount{
					{
						Source: fmt.Sprintf("%s/", p),
						Target: p,
					},
				},
				Entrypoint: []string{"/bin/sh"},
				Args: []string{
					"-c",
					fmt.Sprintf("grep baz %s && tail -f /dev/null", f),
				},
			},
		},
		ConfigFiles: map[string]string{},
	}

	h.ConfigFiles[f] = "baz"

	hcc, err := h.New()
	if err != nil {
		t.Fatalf("Initializing host configured container should succeed, got: %v", err)
	}

	if err = hcc.Configure([]string{f}); err != nil {
		t.Fatalf("Configuring host configured container should succeed, got: %v", err)
	}

	if err = hcc.Create(); err != nil {
		t.Fatalf("Creating host configured container should succeed, got: %v", err)
	}

	if err = hcc.Start(); err != nil {
		t.Fatalf("Starting host configured container should succeed, got: %v", err)
	}

	// Sleep a bit, to make sure container starts etc.
	time.Sleep(containerRunningDelay)

	if err = hcc.Status(); err != nil {
		t.Fatalf("Checking host configured container status should succeed, got: %v", err)
	}

	s := hcc.(*hostConfiguredContainer).container.Status().Status
	if s != "running" {
		t.Errorf("Host configured container should be running, got status %v", s)
	}

	if err = hcc.Stop(); err != nil {
		t.Errorf("Stopping host configured container status should succeed, got: %v", err)
	}

	if err = hcc.Delete(); err != nil {
		t.Fatalf("Deleting host configured container status should succeed, got: %v", err)
	}
}

func TestHostConfiguredContainerPostStartHook(t *testing.T) {
	t.Parallel()

	hookCalled := false

	f := Hook(func() error {
		hookCalled = true

		return nil
	})

	h := &HostConfiguredContainer{
		Host: host.Host{
			DirectConfig: &direct.Config{},
		},
		Container: Container{
			Runtime: RuntimeConfig{
				Docker: &docker.Config{},
			},
			Config: types.ContainerConfig{
				Name:  "foo",
				Image: "busybox:latest",
			},
		},
		Hooks: &Hooks{
			PostStart: &f,
		},
	}

	hcc, err := h.New()
	if err != nil {
		t.Fatalf("Initializing host configured container should succeed, got: %v", err)
	}

	if err = hcc.Create(); err != nil {
		t.Fatalf("Creating host configured container should succeed, got: %v", err)
	}

	if err = hcc.Start(); err != nil {
		t.Fatalf("Starting host configured container should succeed, got: %v", err)
	}

	if !hookCalled {
		t.Errorf("PostStart hook should be called")
	}

	if err = hcc.Stop(); err != nil {
		t.Errorf("Stopping host configured container status should succeed, got: %v", err)
	}

	if err = hcc.Delete(); err != nil {
		t.Fatalf("Deleting host configured container status should succeed, got: %v", err)
	}
}
