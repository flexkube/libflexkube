//go:build integration
// +build integration

package docker

import (
	"reflect"
	"testing"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/flexkube/libflexkube/pkg/container/runtime"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/defaults"
)

// Create() tests.
func TestContainerCreate(t *testing.T) {
	t.Parallel()

	testRuntime, _ := getDockerRuntime(t)

	cc := &types.ContainerConfig{
		Image: defaults.EtcdImage,
	}

	createdContainerID, err := testRuntime.Create(cc)
	if err != nil {
		t.Errorf("Creating container should succeed, got: %s", err)
	}

	t.Cleanup(func() {
		if err := testRuntime.Delete(createdContainerID); err != nil {
			t.Logf("Removing container should succeed, got: %v", err)
		}
	})
}

func TestContainerCreateDelete(t *testing.T) {
	t.Parallel()

	testRuntime, _ := getDockerRuntime(t)

	cc := &types.ContainerConfig{
		Image: defaults.EtcdImage,
	}

	createdContainerID, err := testRuntime.Create(cc)
	if err != nil {
		t.Fatalf("Creating container should succeed, got: %s", err)
	}

	if err := testRuntime.Delete(createdContainerID); err != nil {
		t.Errorf("Removing container should succeed, got: %s", err)
	}
}

func TestContainerCreateNonExistingImage(t *testing.T) {
	t.Parallel()

	testRuntime, _ := getDockerRuntime(t)

	cc := &types.ContainerConfig{
		Image: "nonexistingimage",
	}

	if _, err := testRuntime.Create(cc); err == nil {
		t.Errorf("Creating container with non-existing image should fail")
	}
}

func TestContainerCreatePullImage(t *testing.T) {
	t.Parallel()

	// Don't use default version of image, to have better chance it can be removed
	image := "gcr.io/etcd-development/etcd:v3.3.0"

	testRuntime, _ := getDockerRuntime(t)

	deleteImage(t, image)

	c := &types.ContainerConfig{
		Image: image,
	}

	createdContainerID, err := testRuntime.Create(c)
	if err != nil {
		t.Fatalf("Creating container should pull image and succeed, got: %s", err)
	}

	if err := testRuntime.Delete(createdContainerID); err != nil {
		t.Errorf("Removing container should succeed, got: %s", err)
	}
}

func TestContainerCreateWithArgs(t *testing.T) {
	t.Parallel()

	args := []string{"--logger=zap"}

	testRuntime, testDocker := getDockerRuntime(t)

	containerConfig := &types.ContainerConfig{
		Image:      defaults.EtcdImage,
		Args:       args,
		Entrypoint: []string{"/usr/local/bin/etcd"},
	}

	createdContainerID, err := testRuntime.Create(containerConfig)
	if err != nil {
		t.Fatalf("Creating container with args should succeed, got: %v", err)
	}

	data, err := testDocker.cli.ContainerInspect(testDocker.ctx, createdContainerID)
	if err != nil {
		t.Fatalf("Inspecting created container should succeed, got: %v", err)
	}

	if !reflect.DeepEqual(data.Args, args) {
		t.Fatalf("Container created with args set should have args set\nExpected: %+v\nGot: %+v\n", args, data.Args)
	}

	t.Cleanup(func() {
		if err := testRuntime.Delete(createdContainerID); err != nil {
			t.Logf("Removing container should succeed, got: %v", err)
		}
	})
}

func TestContainerCreateWithEntrypoint(t *testing.T) {
	t.Parallel()

	entrypoint := []string{"/bin/bash"}

	testRuntime, testDocker := getDockerRuntime(t)

	c := &types.ContainerConfig{
		Image:      defaults.EtcdImage,
		Entrypoint: entrypoint,
	}

	createdContainerID, err := testRuntime.Create(c)
	if err != nil {
		t.Fatalf("Creating container with entrypoint should succeed, got: %v", err)
	}

	data, err := testDocker.cli.ContainerInspect(testDocker.ctx, createdContainerID)
	if err != nil {
		t.Fatalf("Inspecting created container should succeed, got: %v", err)
	}

	if !reflect.DeepEqual(data.Path, entrypoint[0]) {
		t.Fatalf("Container created with entrypoint set should have entrypoint set\n"+
			"Expected: %+v\nGot: %+v\n", entrypoint[0], data.Path)
	}

	t.Cleanup(func() {
		if err := testRuntime.Delete(createdContainerID); err != nil {
			t.Logf("Removing container should succeed, got: %v", err)
		}
	})
}

// Start() tests.
func TestContainerStart(t *testing.T) {
	t.Parallel()

	testRuntime, _ := getDockerRuntime(t)

	c := &types.ContainerConfig{
		Image: defaults.EtcdImage,
	}

	createdContainerID, err := testRuntime.Create(c)
	if err != nil {
		t.Fatalf("Creating container should succeed, got: %s", err)
	}

	if err := testRuntime.Start(createdContainerID); err != nil {
		t.Errorf("Starting container should work, got: %s", err)
	}

	t.Cleanup(func() {
		if err := testRuntime.Stop(createdContainerID); err != nil {
			t.Logf("Stopping container should succeed, got: %v", err)

			// Deleting not stopped container will fail, so return early.
			return
		}

		if err := testRuntime.Delete(createdContainerID); err != nil {
			t.Logf("Removing container should succeed, got: %v", err)
		}
	})
}

// Stop() tests.
func TestContainerStop(t *testing.T) {
	t.Parallel()

	testRuntime, _ := getDockerRuntime(t)

	c := &types.ContainerConfig{
		Image: defaults.EtcdImage,
	}

	createdContainerID, err := testRuntime.Create(c)
	if err != nil {
		t.Fatalf("Creating container should succeed, got: %s", err)
	}

	if err := testRuntime.Start(createdContainerID); err != nil {
		t.Fatalf("Starting container should work, got: %s", err)
	}

	if err := testRuntime.Stop(createdContainerID); err != nil {
		t.Errorf("Stopping container should work, got: %s", err)
	}

	t.Cleanup(func() {
		if err := testRuntime.Delete(createdContainerID); err != nil {
			t.Logf("Removing container should succeed, got: %v", err)
		}
	})
}

// Status() tests.
func TestContainerStatus(t *testing.T) {
	t.Parallel()

	testRuntime, _ := getDockerRuntime(t)

	c := &types.ContainerConfig{
		Image: defaults.EtcdImage,
	}

	createdContainerID, err := testRuntime.Create(c)
	if err != nil {
		t.Errorf("Creating container should succeed, got: %s", err)
	}

	if _, err = testRuntime.Status(createdContainerID); err != nil {
		t.Errorf("Getting container status should work, got: %s", err)
	}

	t.Cleanup(func() {
		if err := testRuntime.Delete(createdContainerID); err != nil {
			t.Logf("Removing container should succeed, got: %v", err)
		}
	})
}

func TestContainerStatusNonExistent(t *testing.T) {
	t.Parallel()

	testRuntime, _ := getDockerRuntime(t)

	status, err := testRuntime.Status("nonexistent")
	if err != nil {
		t.Errorf("Getting non-existent container status shouldn't return error, got: %s", err)
	}

	if status.ID != "" {
		t.Errorf("Getting non-existent container status shouldn't return any status")
	}
}

func getDockerRuntime(t *testing.T) (runtime.Runtime, *docker) {
	t.Helper()

	dc := &Config{}

	dockerClient, err := dc.New()
	if err != nil {
		t.Fatalf("Creating new docker runtime should succeed, got: %s", err)
	}

	dockerRuntime, ok := dockerClient.(*docker)
	if !ok {
		t.Fatalf("Unexpected runtime type %T", dockerClient)
	}

	return dockerClient, dockerRuntime
}

func getDockerClient(t *testing.T) *client.Client {
	t.Helper()

	internalDockerClient, err := (&Config{}).getDockerClient()
	if err != nil {
		t.Fatalf("Failed creating Docker client: %v", err)
	}

	dockerClient, ok := internalDockerClient.(*client.Client)
	if !ok {
		t.Fatalf("Got unexpected docker client type %T", internalDockerClient)
	}

	return dockerClient
}

func deleteImage(t *testing.T, image string) {
	t.Helper()

	_, testDocker := getDockerRuntime(t)

	imageID, err := testDocker.imageID(image)
	if err != nil {
		t.Fatalf("Finding image to delete failed: %v", err)
	}

	if imageID == "" {
		return
	}

	c := getDockerClient(t)

	if _, err := c.ImageRemove(testDocker.ctx, imageID, dockertypes.ImageRemoveOptions{}); err != nil {
		t.Fatalf("Removing existing docker image should succeed, got: %v", err)
	}
}

// imageID() tests.
func TestImageID(t *testing.T) {
	t.Parallel()

	_, testDocker := getDockerRuntime(t)

	image := "haproxy:2.0.7-alpine"

	// Make sure image is present on the host.
	if err := testDocker.pullImage(image); err != nil {
		t.Fatalf("Pulling image failed: %v", err)
	}

	id, err := testDocker.imageID(image)
	if err != nil {
		t.Fatalf("Checking image presence failed: %v", err)
	}

	if id == "" {
		t.Fatalf("Pre-pulled image should be present")
	}
}

func TestImageIDMissing(t *testing.T) {
	t.Parallel()

	_, testDocker := getDockerRuntime(t)

	image := "wrk2:latest"

	deleteImage(t, image)

	id, err := testDocker.imageID(image)
	if err != nil {
		t.Fatalf("Getting image ID failed: %v", err)
	}

	if id != "" {
		t.Fatalf("Deleted image should not be not found")
	}
}

// pullImage() tests.
func TestPullImage(t *testing.T) {
	t.Parallel()

	_, testDocker := getDockerRuntime(t)

	image := "busybox:latest"

	deleteImage(t, image)

	imageID, err := testDocker.imageID(image)
	if err != nil {
		t.Fatalf("Getting image ID failed: %v", err)
	}

	if imageID != "" {
		t.Fatalf("Deleted image should not be not found")
	}

	if err := testDocker.pullImage(image); err != nil {
		t.Fatalf("Pulling image failed: %v", err)
	}

	imageID, err = testDocker.imageID(image)
	if err != nil {
		t.Fatalf("Getting image ID failed: %v", err)
	}

	if imageID == "" {
		t.Fatalf("Pulled image should be present")
	}
}

func TestContainerEnv(t *testing.T) {
	t.Parallel()

	env := map[string]string{"foo": "bar"}
	envSlice := []string{"foo=bar", "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"}

	testRuntime, testDocker := getDockerRuntime(t)

	c := &types.ContainerConfig{
		Image: defaults.EtcdImage,
		Env:   env,
	}

	createdContainerID, err := testRuntime.Create(c)
	if err != nil {
		t.Fatalf("Creating container with environment variables should succeed, got: %v", err)
	}

	data, err := testDocker.cli.ContainerInspect(testDocker.ctx, createdContainerID)
	if err != nil {
		t.Fatalf("Inspecting created container should succeed, got: %v", err)
	}

	if !reflect.DeepEqual(data.Config.Env, envSlice) {
		t.Fatalf("Container created with environment variables set should have environment variables set"+
			"\nExpected: %+v\nGot: %+v\n", envSlice, data.Config.Env)
	}

	t.Cleanup(func() {
		if err := testRuntime.Delete(createdContainerID); err != nil {
			t.Logf("Removing container should succeed, got: %v", err)
		}
	})
}
