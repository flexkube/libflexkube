//go:build integration
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
	// Arbitrary value of wow long we wait for container to start and report as running by Docker.
	containerRunningDelay = 3 * time.Second
)

// Create() tests.
//
//nolint:funlen // Just a long integration test.
func TestHostConfiguredContainerDeployConfigFile(t *testing.T) {
	t.Parallel()

	basePath := "/tmp/foo"
	filePath := path.Join(basePath, randomContainerName(t))

	hccConfig := &HostConfiguredContainer{
		Host: host.Host{
			DirectConfig: &direct.Config{},
		},
		Container: Container{
			Runtime: RuntimeConfig{
				Docker: &docker.Config{},
			},
			Config: types.ContainerConfig{
				Name:  randomContainerName(t),
				Image: "nginx",
				Mounts: []types.Mount{
					{
						Source: fmt.Sprintf("%s/", basePath),
						Target: basePath,
					},
				},
				Entrypoint: []string{"/bin/sh"},
				Args: []string{
					"-c",
					fmt.Sprintf("grep baz %s && tail -f /dev/null", filePath),
				},
			},
		},
		ConfigFiles: map[string]string{
			filePath: "baz",
		},
	}

	hcc, err := hccConfig.New()
	if err != nil {
		t.Fatalf("Initializing host configured container should succeed, got: %v", err)
	}

	if err = hcc.Configure([]string{filePath}); err != nil {
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

	testHCC, ok := hcc.(*hostConfiguredContainer)
	if !ok {
		t.Fatalf("Unexpected type for host configured container: %T", hcc)
	}

	s := testHCC.container.Status().Status
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

	hookF := Hook(func() error {
		hookCalled = true

		return nil
	})

	hccConfig := &HostConfiguredContainer{
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
			PostStart: &hookF,
		},
	}

	hcc, err := hccConfig.New()
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
