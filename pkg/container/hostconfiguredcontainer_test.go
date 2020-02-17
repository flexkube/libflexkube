package container

import (
	"fmt"
	"net"
	"testing"

	"github.com/flexkube/libflexkube/pkg/container/runtime"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

// withHook()
func TestWithHook(t *testing.T) {
	action := false

	if err := withHook(nil, func() error {
		action = true

		return nil
	}, nil); err != nil {
		t.Fatalf("withHook should not return error, got: %v", err)
	}

	if !action {
		t.Fatalf("withHook should execute action")
	}
}

func TestWithPreHook(t *testing.T) {
	pre := false

	f := Hook(func() error {
		pre = true

		return nil
	})

	if err := withHook(&f, func() error {
		return nil
	}, nil); err != nil {
		t.Fatalf("withHook should not return error, got: %v", err)
	}

	if !pre {
		t.Fatalf("withHook should call pre-hook")
	}
}

func TestWithPostHook(t *testing.T) {
	post := false

	f := Hook(func() error {
		post = true

		return nil
	})

	if err := withHook(nil, func() error {
		return nil
	}, &f); err != nil {
		t.Fatalf("withHook should not return error, got: %v", err)
	}

	if !post {
		t.Fatalf("withHook should call post-hook")
	}
}

func TestConnectAndForward(t *testing.T) {
	addr := &net.UnixAddr{
		Name: "@foo",
		Net:  "unix",
	}

	localSock, err := net.ListenUnix("unix", addr)
	if err != nil {
		t.Fatalf("unable to listen on address '%s':%v", addr, err)
	}
	defer localSock.Close()

	h := &hostConfiguredContainer{
		host: host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	s, err := h.connectAndForward(fmt.Sprintf("unix://%s", addr.String()))
	if err != nil {
		t.Fatalf("Direct forwarding to open listener should work, got: %v", err)
	}

	if s == "" {
		t.Fatalf("Returned forwarded address shouldn't be empty")
	}
}

// Status()
func TestHostConfiguredContainerStatusNotExist(t *testing.T) {
	h := &hostConfiguredContainer{
		container: &container{},
	}

	if err := h.Status(); err != nil {
		t.Fatalf("checking status of non existing container should succeed, got: %v", err)
	}
}

func TestHostConfiguredContainerStatus(t *testing.T) {
	h := &hostConfiguredContainer{
		host: host.Host{
			DirectConfig: &direct.Config{},
		},
		container: &container{
			base: base{
				runtimeConfig: &runtime.FakeConfig{
					Runtime: &runtime.Fake{
						StatusF: func(id string) (types.ContainerStatus, error) {
							return types.ContainerStatus{}, nil
						},
					},
				},
				status: types.ContainerStatus{
					ID: "foo",
				},
			},
		},
	}

	if err := h.Status(); err != nil {
		t.Fatalf("checking status of existing container should succeed, got: %v", err)
	}
}

// createConfigurationContainer()
func TestHostConfiguredContainerCreateConfigurationContainer(t *testing.T) {
	h := &hostConfiguredContainer{
		container: &container{
			base: base{
				runtime: &runtime.Fake{
					CreateF: func(config *types.ContainerConfig) (string, error) {
						return "", fmt.Errorf("creating failed")
					},
				},
			},
		},
	}

	if err := h.createConfigurationContainer(); err == nil {
		t.Fatalf("creating configuration container should fail")
	}
}

// removeConfigurationContainer()
func TestHostConfiguredContainerRemoveConfigurationContainer(t *testing.T) {
	deleted := false
	i := "foo"

	h := &hostConfiguredContainer{
		configContainer: &containerInstance{
			base: base{
				runtime: &runtime.Fake{
					StatusF: func(id string) (types.ContainerStatus, error) {
						return types.ContainerStatus{
							ID: i,
						}, nil
					},
					DeleteF: func(id string) error {
						if id != i {
							t.Fatalf("should remove container %s, got %s", i, id)
						}

						deleted = true

						return nil
					},
				},
				status: types.ContainerStatus{
					ID: i,
				},
			},
		},
	}

	if err := h.removeConfigurationContainer(); err != nil {
		t.Fatalf("removing configuration container should succeed, got: %v", err)
	}

	if !deleted {
		t.Fatalf("removing existing configuration container should call Delete()")
	}
}

func TestHostConfiguredContainerRemoveConfigurationContainerFailStatus(t *testing.T) {
	h := &hostConfiguredContainer{
		configContainer: &containerInstance{
			base: base{
				runtime: &runtime.Fake{
					StatusF: func(id string) (types.ContainerStatus, error) {
						return types.ContainerStatus{}, fmt.Errorf("checking status failed")
					},
				},
			},
		},
	}

	if err := h.removeConfigurationContainer(); err == nil {
		t.Fatalf("removing configuration container should fail")
	}
}
