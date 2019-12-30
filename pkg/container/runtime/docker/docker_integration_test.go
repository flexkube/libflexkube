// +build integration

package docker

import (
	"reflect"
	"testing"

	dockertypes "github.com/docker/docker/api/types"

	"github.com/flexkube/libflexkube/pkg/container/runtime"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/defaults"
)

// Create
func TestContainerCreate(t *testing.T) {
	r, _ := getDockerRuntime(t)

	cc := &types.ContainerConfig{
		Image: defaults.EtcdImage,
	}

	if _, err := r.Create(cc); err != nil {
		t.Errorf("Creating container should succeed, got: %s", err)
	}
}

func TestContainerCreateDelete(t *testing.T) {
	r, _ := getDockerRuntime(t)

	cc := &types.ContainerConfig{
		Image: defaults.EtcdImage,
	}

	id, err := r.Create(cc)
	if err != nil {
		t.Fatalf("Creating container should succeed, got: %s", err)
	}

	if err := r.Delete(id); err != nil {
		t.Errorf("Removing container should succeed, got: %s", err)
	}
}

func TestContainerCreateNonExistingImage(t *testing.T) {
	r, _ := getDockerRuntime(t)

	cc := &types.ContainerConfig{
		Image: "nonexistingimage",
	}

	if _, err := r.Create(cc); err == nil {
		t.Errorf("Creating container with non-existing image should fail")
	}
}

func TestContainerCreatePullImage(t *testing.T) {
	// Don't use default version of image, to have better chance it can be removed
	image := "gcr.io/etcd-development/etcd:v3.3.0"

	r, d := getDockerRuntime(t)

	images, err := d.cli.ImageList(d.ctx, dockertypes.ImageListOptions{})
	if err != nil {
		t.Fatalf("Listing docker images should succeed, got: %v", err)
	}

	for _, i := range images {
		for _, tag := range i.RepoTags {
			if tag == image {
				if _, err := d.cli.ImageRemove(d.ctx, i.ID, dockertypes.ImageRemoveOptions{}); err != nil {
					t.Fatalf("Removing existing docker image should succeed, got: %v", err)
				}
			}
		}
	}

	c := &types.ContainerConfig{
		Image: image,
	}

	id, err := r.Create(c)
	if err != nil {
		t.Fatalf("Creating container should pull image and succeed, got: %s", err)
	}

	if err := r.Delete(id); err != nil {
		t.Errorf("Removing container should succeed, got: %s", err)
	}
}

func TestContainerCreateWithArgs(t *testing.T) {
	args := []string{"--logger=zap"}

	r, d := getDockerRuntime(t)

	c := &types.ContainerConfig{
		Image:      defaults.EtcdImage,
		Args:       args,
		Entrypoint: []string{"/usr/local/bin/etcd"},
	}

	id, err := r.Create(c)
	if err != nil {
		t.Fatalf("Creating container with args should succeed, got: %v", err)
	}

	data, err := d.cli.ContainerInspect(d.ctx, id)
	if err != nil {
		t.Fatalf("Inspecting created container should succeed, got: %v", err)
	}

	if !reflect.DeepEqual(data.Args, args) {
		t.Fatalf("Container created with args set should have args set\nExpected: %+v\nGot: %+v\n", args, data.Args)
	}
}

func TestContainerCreateWithEntrypoint(t *testing.T) {
	entrypoint := []string{"/bin/bash"}

	r, d := getDockerRuntime(t)

	c := &types.ContainerConfig{
		Image:      defaults.EtcdImage,
		Entrypoint: entrypoint,
	}

	id, err := r.Create(c)
	if err != nil {
		t.Fatalf("Creating container with entrypoint should succeed, got: %v", err)
	}

	data, err := d.cli.ContainerInspect(d.ctx, id)
	if err != nil {
		t.Fatalf("Inspecting created container should succeed, got: %v", err)
	}

	if !reflect.DeepEqual(data.Path, entrypoint[0]) {
		t.Fatalf("Container created with entrypoint set should have entrypoint set\nExpected: %+v\nGot: %+v\n", entrypoint[0], data.Path)
	}
}

// Start()
func TestContainerStart(t *testing.T) {
	r, _ := getDockerRuntime(t)

	c := &types.ContainerConfig{
		Image: defaults.EtcdImage,
	}

	id, err := r.Create(c)
	if err != nil {
		t.Fatalf("Creating container should succeed, got: %s", err)
	}

	if err := r.Start(id); err != nil {
		t.Errorf("Starting container should work, got: %s", err)
	}
}

// Stop()
func TestContainerStop(t *testing.T) {
	r, _ := getDockerRuntime(t)

	c := &types.ContainerConfig{
		Image: defaults.EtcdImage,
	}

	id, err := r.Create(c)
	if err != nil {
		t.Fatalf("Creating container should succeed, got: %s", err)
	}

	if err := r.Start(id); err != nil {
		t.Fatalf("Starting container should work, got: %s", err)
	}

	if err := r.Stop(id); err != nil {
		t.Errorf("Stopping container should work, got: %s", err)
	}
}

// Status()
func TestContainerStatus(t *testing.T) {
	r, _ := getDockerRuntime(t)

	c := &types.ContainerConfig{
		Image: defaults.EtcdImage,
	}

	id, err := r.Create(c)
	if err != nil {
		t.Errorf("Creating container should succeed, got: %s", err)
	}

	if _, err = r.Status(id); err != nil {
		t.Errorf("Getting container status should work, got: %s", err)
	}
}

func TestContainerStatusNonExistent(t *testing.T) {
	r, _ := getDockerRuntime(t)

	status, err := r.Status("nonexistent")
	if err != nil {
		t.Errorf("Getting non-existent container status shouldn't return error, got: %s", err)
	}

	if status != nil {
		t.Errorf("Getting non-existent container status shouldn't return any status")
	}
}

func getDockerRuntime(t *testing.T) (runtime.Runtime, *docker) {
	dc := &Config{}

	r, err := dc.New()
	if err != nil {
		t.Fatalf("Creating new docker runtime should succeed, got: %s", err)
	}

	return r, (r.(*docker))
}
