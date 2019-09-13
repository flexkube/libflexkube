// +build integration

package docker

import (
	"testing"

	"github.com/docker/docker/api/types"

	"github.com/invidian/etcd-ariadnes-thread/pkg/container/runtime"
	"github.com/invidian/etcd-ariadnes-thread/pkg/defaults"
)

// Create
func TestContainerCreate(t *testing.T) {
	d, err := New(&Docker{})
	if err != nil {
		t.Errorf("Creating new docker runtime should succeed, got: %s", err)
	}
	c := &runtime.Config{
		Image: defaults.EtcdImage,
	}

	if _, err = d.Create(c); err != nil {
		t.Errorf("Creating container should succeed, got: %s", err)
	}
}

func TestContainerCreateDelete(t *testing.T) {
	d, err := New(&Docker{})
	if err != nil {
		t.Fatalf("Creating new docker runtime should succeed, got: %s", err)
	}
	c := &runtime.Config{
		Image: defaults.EtcdImage,
	}
	id, err := d.Create(c)
	if err != nil {
		t.Fatalf("Creating container should succeed, got: %s", err)
	}

	if err := d.Delete(id); err != nil {
		t.Errorf("Removing container should succeed, got: %s", err)
	}
}

func TestContainerCreateNonExistingImage(t *testing.T) {
	d, err := New(&Docker{})
	if err != nil {
		t.Errorf("Creating new docker runtime should succeed, got: %s", err)
	}
	c := &runtime.Config{
		Image: "nonexistingimage",
	}

	if _, err = d.Create(c); err == nil {
		t.Errorf("Creating container with non-existing image should fail")
	}
}

func TestContainerCreatePullImage(t *testing.T) {
	// Don't use default version of image, to have better chance it can be removed
	image := "gcr.io/etcd-development/etcd:v3.3.0"
	d, err := New(&Docker{})
	if err != nil {
		t.Fatalf("Creating new docker runtime should succeed, got: %s", err)
	}

	images, err := d.cli.ImageList(d.ctx, types.ImageListOptions{})
	if err != nil {
		t.Fatalf("Listing docker images should succeed, got: %w", err)
	}
	for _, i := range images {
		for _, tag := range i.RepoTags {
			if tag == image {
				if _, err := d.cli.ImageRemove(d.ctx, i.ID, types.ImageRemoveOptions{}); err != nil {
					t.Fatalf("Removing existing docker image should succeed, got: %w", err)
				}
			}
		}
	}

	c := &runtime.Config{
		Image: image,
	}
	id, err := d.Create(c)
	if err != nil {
		t.Fatalf("Creating container should pull image and succeed, got: %s", err)
	}

	if err := d.Delete(id); err != nil {
		t.Errorf("Removing container should succeed, got: %s", err)
	}
}

// Start()
func TestContainerStart(t *testing.T) {
	d, err := New(&Docker{})
	if err != nil {
		t.Fatalf("Creating new docker runtime should succeed, got: %s", err)
	}
	c := &runtime.Config{
		Image: defaults.EtcdImage,
	}
	id, err := d.Create(c)
	if err != nil {
		t.Fatalf("Creating container should succeed, got: %s", err)
	}

	if err := d.Start(id); err != nil {
		t.Errorf("Starting container should work, got: %s", err)
	}
}

// Stop()
func TestContainerStop(t *testing.T) {
	d, err := New(&Docker{})
	if err != nil {
		t.Fatalf("Creating new docker runtime should succeed, got: %s", err)
	}
	c := &runtime.Config{
		Image: defaults.EtcdImage,
	}
	id, err := d.Create(c)
	if err != nil {
		t.Fatalf("Creating container should succeed, got: %s", err)
	}
	if err := d.Start(id); err != nil {
		t.Fatalf("Starting container should work, got: %s", err)
	}

	if err := d.Stop(id); err != nil {
		t.Errorf("Stopping container should work, got: %s", err)
	}
}

// Status()
func TestContainerStatus(t *testing.T) {
	d, err := New(&Docker{})
	if err != nil {
		t.Errorf("Creating new docker runtime should succeed, got: %s", err)
	}
	c := &runtime.Config{
		Image: defaults.EtcdImage,
	}
	id, err := d.Create(c)
	if err != nil {
		t.Errorf("Creating container should succeed, got: %s", err)
	}

	if _, err = d.Status(id); err != nil {
		t.Errorf("Getting container status should work, got: %s", err)
	}
}

func TestContainerStatusNonExistent(t *testing.T) {
	d, err := New(&Docker{})
	if err != nil {
		t.Errorf("Creating new docker runtime should succeed, got: %s", err)
	}

	status, err := d.Status("nonexistent")
	if err != nil {
		t.Errorf("Getting non-existent container status shouldn't return error, got: %s", err)
	}
	if status != nil {
		t.Errorf("Getting non-existent container status shouldn't return any status")
	}
}
