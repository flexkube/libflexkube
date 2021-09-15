package container

import (
	"fmt"
	"net"
	"os"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/container/runtime"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

// withHook() tests.
func TestWithHook(t *testing.T) {
	t.Parallel()

	action := false

	if err := withHook(nil, func() error {
		action = true

		return nil
	}, nil); err != nil {
		t.Fatalf("WithHook should not return error, got: %v", err)
	}

	if !action {
		t.Fatalf("WithHook should execute action")
	}
}

func TestWithPreHook(t *testing.T) {
	t.Parallel()

	pre := false

	f := Hook(func() error {
		pre = true

		return nil
	})

	if err := withHook(&f, func() error {
		return nil
	}, nil); err != nil {
		t.Fatalf("WithHook should not return error, got: %v", err)
	}

	if !pre {
		t.Fatalf("WithHook should call pre-hook")
	}
}

func TestWithPostHook(t *testing.T) {
	t.Parallel()

	post := false

	f := Hook(func() error {
		post = true

		return nil
	})

	if err := withHook(nil, func() error {
		return nil
	}, &f); err != nil {
		t.Fatalf("WithHook should not return error, got: %v", err)
	}

	if !post {
		t.Fatalf("WithHook should call post-hook")
	}
}

func TestConnectAndForward(t *testing.T) {
	t.Parallel()

	addr := &net.UnixAddr{
		Name: "@foo",
		Net:  "unix",
	}

	localSock, err := net.ListenUnix("unix", addr)
	if err != nil {
		t.Fatalf("Unable to listen on address %q: %v", addr, err)
	}

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

	if err := localSock.Close(); err != nil {
		t.Logf("Failed to close local socket listener: %v", err)
	}
}

// Status() tests.
func TestHostConfiguredContainerStatusNotExist(t *testing.T) {
	t.Parallel()

	h := &hostConfiguredContainer{
		container: &container{},
	}

	if err := h.Status(); err == nil {
		t.Fatalf("Checking status of non existing container should fail, got: %v", err)
	}
}

func TestHostConfiguredContainerStatus(t *testing.T) {
	t.Parallel()

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
					ID: foo,
				},
			},
		},
	}

	if err := h.Status(); err != nil {
		t.Fatalf("Checking status of existing container should succeed, got: %v", err)
	}
}

// createConfigurationContainer() tests.
func TestHostConfiguredContainerCreateConfigurationContainer(t *testing.T) {
	t.Parallel()

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
		t.Fatalf("Creating configuration container should fail")
	}
}

// removeConfigurationContainer() tests.
func TestHostConfiguredContainerRemoveConfigurationContainer(t *testing.T) {
	t.Parallel()

	deleted := false
	i := foo

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
							t.Fatalf("Should remove container %q, got %q", i, id)
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
		t.Fatalf("Removing configuration container should succeed, got: %v", err)
	}

	if !deleted {
		t.Fatalf("Removing existing configuration container should call Delete()")
	}
}

func TestHostConfiguredContainerRemoveConfigurationContainerFailStatus(t *testing.T) {
	t.Parallel()

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
		t.Fatalf("Removing configuration container should fail")
	}
}

// statMounts() tests.
func TestStatMountsNoMounts(t *testing.T) {
	t.Parallel()

	h := &hostConfiguredContainer{
		container: &container{},
	}

	if _, err := h.statMounts(); err != nil {
		t.Fatalf("Stating mounts when there is no mounts defined should always succeed, got: %v", err)
	}
}

func TestStatMountsRuntimeError(t *testing.T) {
	t.Parallel()

	h := &hostConfiguredContainer{
		configContainer: &containerInstance{
			base: base{
				runtime: &runtime.Fake{
					StatF: func(ID string, paths []string) (map[string]os.FileMode, error) {
						return map[string]os.FileMode{}, fmt.Errorf("stating failed")
					},
				},
			},
		},
		container: &container{
			base{
				config: types.ContainerConfig{
					Mounts: []types.Mount{
						{
							Source: "/etc/",
							Target: "/etc",
						},
					},
				},
			},
		},
	}

	if _, err := h.statMounts(); err == nil {
		t.Fatalf("Stating mount should fail when runtime error occurs")
	}
}

func TestStatMounts(t *testing.T) {
	t.Parallel()

	m := map[string]os.FileMode{
		"/etc": os.ModeDir,
	}

	h := &hostConfiguredContainer{
		configContainer: &containerInstance{
			base: base{
				runtime: &runtime.Fake{
					StatF: func(ID string, paths []string) (map[string]os.FileMode, error) {
						return m, nil
					},
				},
			},
		},
		container: &container{
			base{
				config: types.ContainerConfig{
					Mounts: []types.Mount{
						{
							Source: "/etc/",
							Target: "/etc",
						},
					},
				},
			},
		},
	}

	s, err := h.statMounts()
	if err != nil {
		t.Fatalf("Stating mount should succeed, got: %v", err)
	}

	if diff := cmp.Diff(m, s); diff != "" {
		t.Fatalf("Received stat result differs from expected one: %s", diff)
	}
}

// createMissingMounts() tests.
func TestCreateMissingMountpointsStatFail(t *testing.T) {
	t.Parallel()

	h := &hostConfiguredContainer{
		configContainer: &containerInstance{
			base: base{
				runtime: &runtime.Fake{
					StatF: func(ID string, paths []string) (map[string]os.FileMode, error) {
						return map[string]os.FileMode{}, fmt.Errorf("stat failed")
					},
				},
			},
		},
		container: &container{
			base{
				config: types.ContainerConfig{
					Mounts: []types.Mount{
						{
							Source: "/etc/",
							Target: "/etc",
						},
					},
				},
			},
		},
	}

	if err := h.createMissingMounts(); err == nil {
		t.Fatalf("Creating missing mountpoints should fail when stating mounts fails")
	}
}

func TestCreateMissingMountpointsMountpointFile(t *testing.T) {
	t.Parallel()

	h := &hostConfiguredContainer{
		configContainer: &containerInstance{
			base: base{
				runtime: &runtime.Fake{
					StatF: func(ID string, paths []string) (map[string]os.FileMode, error) {
						return map[string]os.FileMode{
							path.Join(ConfigMountpoint, "/etc"): os.ModePerm,
						}, nil
					},
				},
			},
		},
		container: &container{
			base{
				config: types.ContainerConfig{
					Mounts: []types.Mount{
						{
							Source: "/etc/",
							Target: "/etc",
						},
					},
				},
			},
		},
	}

	if err := h.createMissingMounts(); err == nil {
		t.Fatalf("Creating missing mountpoints should fail when stated mount is a file")
	}
}

func TestCreateMissingMountpointsNoMountsToCreate(t *testing.T) {
	t.Parallel()

	h := &hostConfiguredContainer{
		configContainer: &containerInstance{
			base: base{
				runtime: &runtime.Fake{
					StatF: func(ID string, paths []string) (map[string]os.FileMode, error) {
						return map[string]os.FileMode{
							path.Join(ConfigMountpoint, "/etc/"): os.ModeDir,
						}, nil
					},
				},
			},
		},
		container: &container{
			base{
				config: types.ContainerConfig{
					Mounts: []types.Mount{
						{
							Source: "/etc/",
							Target: "/etc",
						},
					},
				},
			},
		},
	}

	if err := h.createMissingMounts(); err != nil {
		t.Fatalf("Creating missing mountpoints without runtime should succeed, "+
			"if there is no mountpoints to create, got: %v", err)
	}
}

func TestCreateMissingMountpointsCopyFail(t *testing.T) {
	t.Parallel()

	h := &hostConfiguredContainer{
		configContainer: &containerInstance{
			base: base{
				runtime: &runtime.Fake{
					StatF: func(ID string, paths []string) (map[string]os.FileMode, error) {
						return map[string]os.FileMode{}, nil
					},
					CopyF: func(id string, files []*types.File) error {
						return fmt.Errorf("copying failed")
					},
				},
			},
		},
		container: &container{
			base{
				config: types.ContainerConfig{
					Mounts: []types.Mount{
						{
							Source: "/etc/",
							Target: "/etc",
						},
					},
				},
			},
		},
	}

	if err := h.createMissingMounts(); err == nil {
		t.Fatalf("Creating missing mountpoints should fail when copying fails")
	}
}

func TestCreateMissingMountpoints(t *testing.T) {
	t.Parallel()

	called := false

	f := []*types.File{
		{
			Path:    fmt.Sprintf("%s/", path.Join(ConfigMountpoint, "/etc/")),
			Mode:    mountpointDirMode,
			Content: "",
		},
	}

	h := &hostConfiguredContainer{
		configContainer: &containerInstance{
			base: base{
				runtime: &runtime.Fake{
					StatF: func(ID string, paths []string) (map[string]os.FileMode, error) {
						return map[string]os.FileMode{}, nil
					},
					CopyF: func(id string, files []*types.File) error {
						if diff := cmp.Diff(f, files); diff != "" {
							t.Fatalf("Received files for creating differs from expected: %s", diff)
						}

						called = true

						return nil
					},
				},
			},
		},
		container: &container{
			base{
				config: types.ContainerConfig{
					Mounts: []types.Mount{
						{
							Source: "/etc/",
							Target: "/etc",
						},
					},
				},
			},
		},
	}

	if err := h.createMissingMounts(); err != nil {
		t.Fatalf("Creating missing mountpoints should succeed, got: %v", err)
	}

	if !called {
		t.Fatalf("Creating missing mountpoints should call Copy from runtime")
	}
}

// dirMounts() tests.
func TestDirMounts(t *testing.T) {
	t.Parallel()

	m := types.Mount{
		Source: "/etc/",
		Target: "/etc",
	}

	h := &hostConfiguredContainer{
		container: &container{
			base{
				config: types.ContainerConfig{
					Mounts: []types.Mount{
						m,
						{
							Source: "/foo",
							Target: "/bar",
						},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(h.dirMounts(), []types.Mount{m}); diff != "" {
		t.Fatalf("Received wrong dir mounts than expected: %s", diff)
	}
}

// withForwardedRuntime() tests.
func TestWithForwardedRuntimeFailForward(t *testing.T) {
	t.Parallel()

	h := &hostConfiguredContainer{
		container: &container{
			base: base{
				runtimeConfig: &runtime.FakeConfig{
					Runtime: &runtime.Fake{},
				},
			},
		},
	}

	if err := h.withForwardedRuntime(func() error {
		return nil
	}); err == nil {
		t.Fatalf("Should fail with bad host")
	}
}

func TestWithForwardedRuntimeFailRuntime(t *testing.T) {
	t.Parallel()

	h := &hostConfiguredContainer{
		host: host.Host{
			DirectConfig: &direct.Config{},
		},
		container: &container{
			base: base{
				runtimeConfig: &runtime.FakeConfig{},
			},
		},
	}

	if err := h.withForwardedRuntime(func() error {
		return nil
	}); err == nil {
		t.Fatalf("Should fail with bad runtime")
	}
}

func TestWithForwardedRuntime(t *testing.T) {
	t.Parallel()

	r := &runtime.Fake{}

	h := &hostConfiguredContainer{
		host: host.Host{
			DirectConfig: &direct.Config{},
		},
		container: &container{
			base: base{
				runtimeConfig: &runtime.FakeConfig{
					Runtime: r,
				},
			},
		},
	}

	// TODO: Test runtime manipulation here.
	if err := h.withForwardedRuntime(func() error {
		return nil
	}); err != nil {
		t.Fatalf("Should work, got: %v", err)
	}
}

// Create() tests.
func TestHostConfiguredContainerCreateFailMountpoints(t *testing.T) {
	t.Parallel()

	h := &hostConfiguredContainer{
		host: host.Host{
			DirectConfig: &direct.Config{},
		},
		container: &container{
			base{
				runtimeConfig: &runtime.FakeConfig{
					Runtime: &runtime.Fake{
						CreateF: func(config *types.ContainerConfig) (string, error) {
							return foo, nil
						},
						DeleteF: func(id string) error {
							return nil
						},
						StatF: func(ID string, paths []string) (map[string]os.FileMode, error) {
							return map[string]os.FileMode{}, fmt.Errorf("stat failed")
						},
						StatusF: func(id string) (types.ContainerStatus, error) {
							return types.ContainerStatus{}, nil
						},
					},
				},
				config: types.ContainerConfig{
					Mounts: []types.Mount{
						{
							Source: "/etc/",
							Target: "/etc",
						},
					},
				},
			},
		},
	}

	if err := h.Create(); err == nil {
		t.Fatalf("Create with failing stat should fail")
	}
}

func TestHostConfiguredContainerCreateFail(t *testing.T) {
	t.Parallel()

	fail := false

	h := &hostConfiguredContainer{
		host: host.Host{
			DirectConfig: &direct.Config{},
		},
		container: &container{
			base{
				runtimeConfig: &runtime.FakeConfig{
					Runtime: &runtime.Fake{
						CreateF: func(config *types.ContainerConfig) (string, error) {
							if fail {
								return "", fmt.Errorf("2nd create fails")
							}

							fail = true

							return foo, nil
						},
						DeleteF: func(id string) error {
							return nil
						},
						StatF: func(ID string, paths []string) (map[string]os.FileMode, error) {
							return map[string]os.FileMode{
								path.Join(ConfigMountpoint, "/etc/"): os.ModeDir,
							}, nil
						},
						StatusF: func(id string) (types.ContainerStatus, error) {
							return types.ContainerStatus{}, nil
						},
					},
				},
				config: types.ContainerConfig{
					Mounts: []types.Mount{
						{
							Source: "/etc/",
							Target: "/etc",
						},
					},
				},
			},
		},
	}

	if err := h.Create(); err == nil {
		t.Fatalf("Create with failing create from runtime should fail")
	}
}

func TestHostConfiguredContainerCreateFailStatus(t *testing.T) {
	t.Parallel()

	fail := false

	h := &hostConfiguredContainer{
		host: host.Host{
			DirectConfig: &direct.Config{},
		},
		container: &container{
			base{
				runtimeConfig: &runtime.FakeConfig{
					Runtime: &runtime.Fake{
						CreateF: func(config *types.ContainerConfig) (string, error) {
							return foo, nil
						},
						DeleteF: func(id string) error {
							return nil
						},
						StatF: func(ID string, paths []string) (map[string]os.FileMode, error) {
							return map[string]os.FileMode{
								path.Join(ConfigMountpoint, "/etc/"): os.ModeDir,
							}, nil
						},
						StatusF: func(id string) (types.ContainerStatus, error) {
							if fail {
								return types.ContainerStatus{}, fmt.Errorf("2nd status fails")
							}

							fail = true

							return types.ContainerStatus{}, nil
						},
					},
				},
				config: types.ContainerConfig{
					Mounts: []types.Mount{
						{
							Source: "/etc/",
							Target: "/etc",
						},
					},
				},
			},
		},
	}

	if err := h.Create(); err == nil {
		t.Fatalf("Create with failing status from runtime should fail")
	}
}

func TestHostConfiguredContainerCreate(t *testing.T) {
	t.Parallel()

	h := &hostConfiguredContainer{
		host: host.Host{
			DirectConfig: &direct.Config{},
		},
		container: &container{
			base{
				runtimeConfig: &runtime.FakeConfig{
					Runtime: &runtime.Fake{
						CreateF: func(config *types.ContainerConfig) (string, error) {
							return foo, nil
						},
						DeleteF: func(id string) error {
							return nil
						},
						StatF: func(ID string, paths []string) (map[string]os.FileMode, error) {
							return map[string]os.FileMode{
								path.Join(ConfigMountpoint, "/etc/"): os.ModeDir,
							}, nil
						},
						StatusF: func(id string) (types.ContainerStatus, error) {
							return types.ContainerStatus{
								ID: "bar",
							}, nil
						},
					},
				},
				config: types.ContainerConfig{
					Mounts: []types.Mount{
						{
							Source: "/etc/",
							Target: "/etc",
						},
					},
				},
			},
		},
	}

	if err := h.Create(); err != nil {
		t.Fatalf("Create should succeed, got: %v", err)
	}

	if id := h.container.Status().ID; id != "bar" {
		t.Fatalf("Expected ID %q, got %q", "bar", id)
	}
}

// updateConfigurationStatus() tests.
func TestHostConfiguredContainerUpdateConfigurationStatusNoAction(t *testing.T) {
	t.Parallel()

	h := &hostConfiguredContainer{
		host: host.Host{
			DirectConfig: &direct.Config{},
		},
		container: &container{
			base{
				runtimeConfig: &runtime.FakeConfig{
					Runtime: &runtime.Fake{
						CreateF: func(config *types.ContainerConfig) (string, error) {
							return foo, nil
						},
						DeleteF: func(id string) error {
							return nil
						},
					},
				},
			},
		},
	}

	if err := h.updateConfigurationStatus(); err != nil {
		t.Fatalf("Updating configuration status without configuration files should always succeed, got: %v", err)
	}
}

func TestHostConfiguredContainerUpdateConfigurationStatusFileMissing(t *testing.T) {
	t.Parallel()

	h := &hostConfiguredContainer{
		configFiles: map[string]string{
			"/foo": "bar",
		},
		host: host.Host{
			DirectConfig: &direct.Config{},
		},
		configContainer: &containerInstance{
			base{
				runtime: &runtime.Fake{
					CreateF: func(config *types.ContainerConfig) (string, error) {
						return foo, nil
					},
					DeleteF: func(id string) error {
						return nil
					},
					ReadF: func(id string, srcPath []string) ([]*types.File, error) {
						if diff := cmp.Diff(srcPath, []string{path.Join(ConfigMountpoint, "/foo")}); diff != "" {
							t.Fatalf("Unexpected srcPath: %s", diff)
						}

						return []*types.File{}, nil
					},
				},
			},
		},
	}

	if err := h.updateConfigurationStatus(); err != nil {
		t.Fatalf("Updating configuration status without configuration files should always succeed, got: %v", err)
	}

	if diff := cmp.Diff(h.configFiles, map[string]string{}); diff != "" {
		t.Fatalf("Updating configuration status should reset configFiles map if no files were found, got: %s", diff)
	}
}

func TestHostConfiguredContainerUpdateConfigurationStatusNewContent(t *testing.T) {
	t.Parallel()

	h := &hostConfiguredContainer{
		configFiles: map[string]string{
			"/foo": "bar",
		},
		host: host.Host{
			DirectConfig: &direct.Config{},
		},
		configContainer: &containerInstance{
			base{
				runtime: &runtime.Fake{
					CreateF: func(config *types.ContainerConfig) (string, error) {
						return foo, nil
					},
					DeleteF: func(id string) error {
						return nil
					},
					ReadF: func(id string, srcPath []string) ([]*types.File, error) {
						return []*types.File{
							{
								Path:    path.Join(ConfigMountpoint, "/foo"),
								Content: "doh",
							},
						}, nil
					},
				},
			},
		},
	}

	if err := h.updateConfigurationStatus(); err != nil {
		t.Fatalf("Updating configuration status without configuration files should always succeed, got: %v", err)
	}

	e := map[string]string{
		"/foo": "doh",
	}

	if diff := cmp.Diff(h.configFiles, e); diff != "" {
		t.Fatalf("Updating configuration status should update content of the file with one returned by runtime: %s", diff)
	}
}

func TestHostConfiguredContainerUpdateConfigurationStatusReadRuntimeError(t *testing.T) {
	t.Parallel()

	h := &hostConfiguredContainer{
		configFiles: map[string]string{
			"/foo": "bar",
		},
		host: host.Host{
			DirectConfig: &direct.Config{},
		},
		configContainer: &containerInstance{
			base{
				runtime: &runtime.Fake{
					CreateF: func(config *types.ContainerConfig) (string, error) {
						return foo, nil
					},
					DeleteF: func(id string) error {
						return nil
					},
					ReadF: func(id string, srcPath []string) ([]*types.File, error) {
						return []*types.File{}, fmt.Errorf("reading")
					},
				},
			},
		},
	}

	if err := h.updateConfigurationStatus(); err == nil {
		t.Fatalf("Updating configuration status should return error when runtime read fails")
	}
}
