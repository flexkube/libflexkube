package container

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/container/runtime"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

const (
	foo = "foo"
	bar = "bar"
)

// New()
func TestContainersNew(t *testing.T) {
	GetContainers(t)
}

func GetContainers(t *testing.T) ContainersInterface {
	cc := &Containers{
		DesiredState: ContainersState{
			foo: &HostConfiguredContainer{
				Host: host.Host{
					DirectConfig: &direct.Config{},
				},
				Container: Container{
					Runtime: RuntimeConfig{
						Docker: &docker.Config{},
					},
					Config: types.ContainerConfig{
						Name:  foo,
						Image: "busybox:latest",
					},
				},
			},
		},
	}

	c, err := cc.New()
	if err != nil {
		t.Fatalf("Creating empty containers object should work, got: %v", err)
	}

	return c
}

// Containers()
func TestContainersContainers(t *testing.T) {
	c := &containers{}

	if !reflect.DeepEqual(c, c.Containers()) {
		t.Fatalf("Containers() should return self")
	}
}

// CheckCurrentState()
func TestContainersCheckCurrentStateNew(t *testing.T) {
	c := GetContainers(t)

	if err := c.CheckCurrentState(); err != nil {
		t.Fatalf("Checking current state for new Containers should work, got: %v", err)
	}
}

// CurrentStateToYaml()
func TestContainersCurrentStateToYAML(t *testing.T) {
	c := GetContainers(t)

	_, err := c.StateToYaml()
	if err != nil {
		t.Fatalf("Getting current state in YAML format should work, got: %v", err)
	}
}

// ToExported()
func TestContainersToExported(t *testing.T) {
	c := GetContainers(t)

	c.ToExported()
}

// FromYaml()
func TestContainersFromYamlBad(t *testing.T) {
	if _, err := FromYaml([]byte{}); err == nil {
		t.Fatalf("Creating containers from empty YAML should fail")
	}
}

func TestContainersFromYaml(t *testing.T) {
	y := `
desiredState:
 foo:
   host:
     direct: {}
   container:
     runtime:
       docker: {}
     config:
       name: foo
       image: busybox
`

	if _, err := FromYaml([]byte(y)); err != nil {
		t.Fatalf("Creating containers from valid YAML should work, got: %v", err)
	}
}

// filesToUpdate
func TestFilesToUpdateEmpty(t *testing.T) {
	expected := []string{foo}

	d := hostConfiguredContainer{
		configFiles: map[string]string{
			foo: bar,
		},
	}

	if v := filesToUpdate(d, nil); !reflect.DeepEqual(expected, v) {
		t.Fatalf("Expected %v, got %v", expected, d)
	}
}

// Validate()
func TestValidateEmpty(t *testing.T) {
	cc := &Containers{}

	if err := cc.Validate(); err == nil {
		t.Fatalf("Empty containers object shouldn't be valid")
	}
}

func TestValidateNil(t *testing.T) {
	var cc Containers

	if err := cc.Validate(); err == nil {
		t.Fatalf("nil containers object shouldn't be valid")
	}
}

func TestValidateNoContainers(t *testing.T) {
	cc := &Containers{
		DesiredState:  ContainersState{},
		PreviousState: ContainersState{},
	}

	if err := cc.Validate(); err == nil {
		t.Fatalf("Containers object without any containers shouldn't be valid")
	}
}

func TestValidateBadDesiredContainers(t *testing.T) {
	cc := &Containers{
		DesiredState: ContainersState{
			foo: &HostConfiguredContainer{},
		},
		PreviousState: ContainersState{},
	}

	if err := cc.Validate(); err == nil {
		t.Fatalf("Containers object with bad desired container shouldn't be valid")
	}
}

func TestValidateBadCurrentContainers(t *testing.T) {
	cc := &Containers{
		DesiredState: ContainersState{},
		PreviousState: ContainersState{
			foo: &HostConfiguredContainer{},
		},
	}

	if err := cc.Validate(); err == nil {
		t.Fatalf("Containers object with bad current container shouldn't be valid")
	}
}

// isUpdatable()
func TestIsUpdatableWithoutCurrentState(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			foo: &hostConfiguredContainer{},
		},
	}

	if err := c.isUpdatable(foo); err == nil {
		t.Fatalf("Container without current state shouldn't be updatable.")
	}
}

func TestIsUpdatableToBeDeleted(t *testing.T) {
	c := &containers{
		currentState: containersState{
			foo: &hostConfiguredContainer{},
		},
	}

	if err := c.isUpdatable(foo); err == nil {
		t.Fatalf("Container without desired state shouldn't be updatable.")
	}
}

func TestIsUpdatable(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			foo: &hostConfiguredContainer{},
		},
		currentState: containersState{
			foo: &hostConfiguredContainer{},
		},
	}

	if err := c.isUpdatable(foo); err != nil {
		t.Fatalf("Container with current and desired state should be updatable, got: %v", err)
	}
}

// diffHost()
func TestDiffHostNotUpdatable(t *testing.T) {
	c := &containers{
		currentState: containersState{
			foo: &hostConfiguredContainer{},
		},
	}

	if _, err := c.diffHost(foo); err == nil {
		t.Fatalf("Not updatable container shouldn't return diff")
	}
}

func TestDiffHostNoDiff(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				host: host.Host{},
			},
		},
		currentState: containersState{
			foo: &hostConfiguredContainer{
				host: host.Host{},
			},
		},
	}

	diff, err := c.diffHost(foo)
	if err != nil {
		t.Fatalf("Updatable container should return diff, got: %v", err)
	}

	if diff != "" {
		t.Fatalf("Container without host updates shouldn't return diff, got: %s", diff)
	}
}

func TestDiffHost(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
			},
		},
		currentState: containersState{
			foo: &hostConfiguredContainer{
				host: host.Host{},
			},
		},
	}

	diff, err := c.diffHost(foo)
	if err != nil {
		t.Fatalf("Updatable container should return diff, got: %v", err)
	}

	if diff == "" {
		t.Fatalf("Container with host updates should return diff")
	}
}

// diffContainer()
func TestDiffContainerNotUpdatable(t *testing.T) {
	c := &containers{
		currentState: containersState{
			foo: &hostConfiguredContainer{},
		},
	}

	if _, err := c.diffContainer(foo); err == nil {
		t.Fatalf("Not updatable container shouldn't return diff")
	}
}

func TestDiffContainerNoDiff(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
			},
		},
		currentState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
			},
		},
	}

	diff, err := c.diffContainer(foo)
	if err != nil {
		t.Fatalf("Updatable container should return diff, got: %v", err)
	}

	if diff != "" {
		t.Fatalf("Container without host updates shouldn't return diff, got: %s", diff)
	}
}

func TestDiffContainer(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
			},
		},
		currentState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{
							Name: foo,
						},
					},
				},
			},
		},
	}

	diff, err := c.diffContainer(foo)
	if err != nil {
		t.Fatalf("Updatable container should return diff, got: %v", err)
	}

	if diff == "" {
		t.Fatalf("Container with config updates should return diff")
	}
}

// ensureRunning()
func TestEnsureRunningNonExistent(t *testing.T) {
	c := &containers{
		currentState: containersState{},
	}

	if err := ensureRunning(c.currentState[bar]); err == nil {
		t.Fatalf("Ensuring that non existing container is running should fail")
	}
}

func TestEnsureRunning(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
			},
		},
		currentState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
						status: types.ContainerStatus{
							ID:     "existing",
							Status: "running",
						},
					},
				},
			},
		},
	}

	if err := ensureRunning(c.currentState[foo]); err != nil {
		t.Fatalf("Ensuring that running container is running should succeed, got: %v", err)
	}
}

// ensureExists()
func TestEnsureExistsAlreadyExists(t *testing.T) {
	c := &containers{
		currentState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
						status: types.ContainerStatus{
							ID: "existing",
						},
					},
				},
			},
		},
	}

	if err := c.ensureExists(foo); err != nil {
		t.Fatalf("Ensuring that existing container exists should succeed, got: %v", err)
	}
}

func TestEnsureExistsFailCreate(t *testing.T) {
	c := &containers{
		currentState: containersState{},
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						config: types.ContainerConfig{},
						runtimeConfig: &runtime.FakeConfig{
							Runtime: &runtime.Fake{
								CreateF: func(config *types.ContainerConfig) (string, error) {
									return "", fmt.Errorf("create fail")
								},
							},
						},
					},
				},
			},
		},
	}

	if err := c.ensureExists(foo); err == nil {
		t.Fatalf("Ensuring that new container exists should propagate create error")
	}
}

func TestEnsureExistsFailStart(t *testing.T) {
	c := &containers{
		currentState: containersState{},
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				hooks: &Hooks{},
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						config: types.ContainerConfig{},
						runtimeConfig: &runtime.FakeConfig{
							Runtime: &runtime.Fake{
								CreateF: func(config *types.ContainerConfig) (string, error) {
									return foo, nil
								},
								StatusF: func(id string) (types.ContainerStatus, error) {
									return types.ContainerStatus{
										ID: bar,
									}, nil
								},
								DeleteF: func(id string) error {
									return nil
								},
								StartF: func(id string) error {
									return fmt.Errorf("start fail")
								},
							},
						},
					},
				},
			},
		},
	}

	if err := c.ensureExists(foo); err == nil {
		t.Fatalf("Ensuring that new container exists should fail")
	}

	if _, ok := c.currentState[foo]; !ok {
		t.Fatalf("ensureExists should save state of created container even if starting failed")
	}
}

func TestEnsureExist(t *testing.T) {
	c := &containers{
		currentState: containersState{},
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				hooks: &Hooks{},
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						config: types.ContainerConfig{},
						runtimeConfig: &runtime.FakeConfig{
							Runtime: &runtime.Fake{
								CreateF: func(config *types.ContainerConfig) (string, error) {
									return foo, nil
								},
								StatusF: func(id string) (types.ContainerStatus, error) {
									return types.ContainerStatus{
										ID: bar,
									}, nil
								},
								DeleteF: func(id string) error {
									return nil
								},
								StartF: func(id string) error {
									return nil
								},
							},
						},
					},
				},
			},
		},
	}

	if err := c.ensureExists(foo); err != nil {
		t.Fatalf("Ensuring that new container exists should succeed, got: %v", err)
	}

	if _, ok := c.currentState[foo]; !ok {
		t.Fatalf("ensureExists should save state of created container when creation succeeds")
	}
}

// ensureHost()
func TestEnsureHostNoDiff(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
			},
		},
		currentState: containersState{
			foo: &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
			},
		},
	}

	if err := c.ensureHost(foo); err != nil {
		t.Fatalf("Ensuring that container's host configuration is up to date should succeed, got: %v", err)
	}
}

func TestEnsureHostFailStart(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				hooks: &Hooks{},
				host: host.Host{
					DirectConfig: &direct.Config{
						Dummy: "foo",
					},
				},
				container: &container{
					base: base{
						runtimeConfig: &runtime.FakeConfig{
							Runtime: &runtime.Fake{
								CreateF: func(config *types.ContainerConfig) (string, error) {
									return foo, nil
								},
								DeleteF: func(id string) error {
									return nil
								},
								StatusF: func(id string) (types.ContainerStatus, error) {
									return types.ContainerStatus{
										ID: bar,
									}, nil
								},
								StartF: func(id string) error {
									return fmt.Errorf("start fails")
								},
							},
						},
					},
				},
			},
		},
		currentState: containersState{
			foo: &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						status: types.ContainerStatus{
							ID: foo,
						},
						runtimeConfig: &runtime.FakeConfig{
							Runtime: &runtime.Fake{
								StatusF: func(id string) (types.ContainerStatus, error) {
									return types.ContainerStatus{
										ID: foo,
									}, nil
								},
								DeleteF: func(id string) error {
									return nil
								},
								StopF: func(id string) error {
									return nil
								},
							},
						},
					},
				},
			},
		},
	}

	if err := c.ensureHost(foo); err == nil {
		t.Fatalf("Ensuring that container's host configuration is up to date should fail")
	}

	if c.currentState[foo].container.Status().ID != bar {
		t.Fatalf("ensure host should persist state changes even if process fails")
	}
}

func TestEnsureHost(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				hooks: &Hooks{},
				host: host.Host{
					DirectConfig: &direct.Config{
						Dummy: "foo",
					},
				},
				container: &container{
					base: base{
						runtimeConfig: &runtime.FakeConfig{
							Runtime: &runtime.Fake{
								CreateF: func(config *types.ContainerConfig) (string, error) {
									return foo, nil
								},
								DeleteF: func(id string) error {
									return nil
								},
								StatusF: func(id string) (types.ContainerStatus, error) {
									return types.ContainerStatus{
										ID: bar,
									}, nil
								},
								StartF: func(id string) error {
									return nil
								},
							},
						},
					},
				},
			},
		},
		currentState: containersState{
			foo: &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						status: types.ContainerStatus{
							ID: foo,
						},
						runtimeConfig: &runtime.FakeConfig{
							Runtime: &runtime.Fake{
								StatusF: func(id string) (types.ContainerStatus, error) {
									return types.ContainerStatus{
										ID: foo,
									}, nil
								},
								DeleteF: func(id string) error {
									return nil
								},
								StopF: func(id string) error {
									return nil
								},
							},
						},
					},
				},
			},
		},
	}

	if err := c.ensureHost(foo); err != nil {
		t.Fatalf("Ensuring that container's host configuration is up to date should succeed, got: %v", err)
	}

	if c.currentState[foo].container.Status().ID != bar {
		t.Fatalf("ensure host should persist state changes even if process fails")
	}
}

// ensureContainer()
func TestEnsureContainerNoDiff(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
			},
		},
		currentState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
			},
		},
	}

	if err := c.ensureContainer(foo); err != nil {
		t.Fatalf("Ensuring that container configuration is up to date should succeed, got: %v", err)
	}
}

func TestEnsureContainerFailStart(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				hooks: &Hooks{},
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						config: types.ContainerConfig{
							Image: foo,
						},
						runtimeConfig: &runtime.FakeConfig{
							Runtime: &runtime.Fake{
								CreateF: func(config *types.ContainerConfig) (string, error) {
									return foo, nil
								},
								DeleteF: func(id string) error {
									return nil
								},
								StatusF: func(id string) (types.ContainerStatus, error) {
									return types.ContainerStatus{
										ID: bar,
									}, nil
								},
								StartF: func(id string) error {
									return fmt.Errorf("start fails")
								},
							},
						},
					},
				},
			},
		},
		currentState: containersState{
			foo: &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						status: types.ContainerStatus{
							ID: foo,
						},
						config: types.ContainerConfig{
							Image: bar,
						},
						runtimeConfig: &runtime.FakeConfig{
							Runtime: &runtime.Fake{
								StatusF: func(id string) (types.ContainerStatus, error) {
									return types.ContainerStatus{
										ID: foo,
									}, nil
								},
								DeleteF: func(id string) error {
									return nil
								},
								StopF: func(id string) error {
									return nil
								},
							},
						},
					},
				},
			},
		},
	}

	if err := c.ensureContainer(foo); err == nil {
		t.Fatalf("Ensuring that container configuration is up to date should fail")
	}

	if c.currentState[foo].container.Status().ID != bar {
		t.Fatalf("ensure container should persist state changes even if process fails")
	}
}

func TestEnsureContainer(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				hooks: &Hooks{},
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						config: types.ContainerConfig{
							Image: foo,
						},
						runtimeConfig: &runtime.FakeConfig{
							Runtime: &runtime.Fake{
								CreateF: func(config *types.ContainerConfig) (string, error) {
									return foo, nil
								},
								DeleteF: func(id string) error {
									return nil
								},
								StatusF: func(id string) (types.ContainerStatus, error) {
									return types.ContainerStatus{
										ID: bar,
									}, nil
								},
								StartF: func(id string) error {
									return nil
								},
							},
						},
					},
				},
			},
		},
		currentState: containersState{
			foo: &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						status: types.ContainerStatus{
							ID: foo,
						},
						config: types.ContainerConfig{
							Image: bar,
						},
						runtimeConfig: &runtime.FakeConfig{
							Runtime: &runtime.Fake{
								StatusF: func(id string) (types.ContainerStatus, error) {
									return types.ContainerStatus{
										ID: foo,
									}, nil
								},
								DeleteF: func(id string) error {
									return nil
								},
								StopF: func(id string) error {
									return nil
								},
							},
						},
					},
				},
			},
		},
	}

	if err := c.ensureContainer(foo); err != nil {
		t.Fatalf("Ensuring that container configuration is up to date should succeed, got: %v", err)
	}

	if c.currentState[foo].container.Status().ID != bar {
		t.Fatalf("ensure container should persist state changes even if process fails")
	}
}

// recreate()
func TestRecreateNonExistent(t *testing.T) {
	c := &containers{}
	if err := c.recreate(foo); err == nil {
		t.Fatalf("Recreating on empty containers should fail")
	}
}

// Deploy()
func TestDeployNoCurrentState(t *testing.T) {
	c := &containers{}
	if err := c.Deploy(); err == nil {
		t.Fatalf("Execute without current state should fail")
	}
}

// hasUpdates()
func TestHasUpdatesHost(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
			},
		},
		currentState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
				host: host.Host{},
			},
		},
	}

	u, err := c.hasUpdates(foo)
	if err != nil {
		t.Fatalf("Checking for updates should succeed, got: %v", err)
	}

	if !u {
		t.Fatalf("Container with host changes should indicate update")
	}
}

func TestHasUpdatesConfig(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
				configFiles: map[string]string{
					foo: foo,
				},
			},
		},
		currentState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
			},
		},
	}

	u, err := c.hasUpdates(foo)
	if err != nil {
		t.Fatalf("Checking for updates should succeed, got: %v", err)
	}

	if !u {
		t.Fatalf("Container with configuration files changes should indicate update")
	}
}

func TestHasUpdatesContainer(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
			},
		},
		currentState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{
							Name: foo,
						},
					},
				},
			},
		},
	}

	u, err := c.hasUpdates(foo)
	if err != nil {
		t.Fatalf("Checking for updates should succeed, got: %v", err)
	}

	if !u {
		t.Fatalf("Container with container configuration changes should indicate update")
	}
}

func TestHasUpdatesNoUpdates(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
			},
		},
		currentState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
			},
		},
	}

	u, err := c.hasUpdates(foo)
	if err != nil {
		t.Fatalf("Checking for updates should succeed, got: %v", err)
	}

	if u {
		t.Fatalf("Container with no changes should indicate no changes")
	}
}

// ensureConfigured()
func TestEnsureConfiguredDisposable(t *testing.T) {
	c := &containers{
		desiredState: containersState{},
		currentState: containersState{
			foo: &hostConfiguredContainer{
				configFiles: map[string]string{},
			},
		},
	}

	if err := c.ensureConfigured(foo); err != nil {
		t.Fatalf("Ensure configured should succeed when container is going to be removed, got: %v", err)
	}
}

func TestEnsureConfiguredNoUpdates(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				configFiles: map[string]string{},
			},
		},
		currentState: containersState{
			foo: &hostConfiguredContainer{
				configFiles: map[string]string{},
			},
		},
	}

	if err := c.ensureConfigured(foo); err != nil {
		t.Fatalf("Ensure configured should succeed, got: %v", err)
	}
}

func TestEnsureConfigured(t *testing.T) {
	called := false

	f := foo
	cf := map[string]string{
		f: bar,
	}

	c := &containers{
		desiredState: containersState{
			f: &hostConfiguredContainer{
				configFiles: cf,
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						config: types.ContainerConfig{
							Image: f,
						},
						runtimeConfig: &runtime.FakeConfig{
							Runtime: &runtime.Fake{
								CreateF: func(config *types.ContainerConfig) (string, error) {
									return f, nil
								},
								StatusF: func(id string) (types.ContainerStatus, error) {
									return types.ContainerStatus{}, nil
								},
								CopyF: func(id string, files []*types.File) error {
									if id != f {
										t.Errorf("should copy to configuration container '%s', not to '%s'", f, id)
									}

									if len(files) != len(cf) {
										t.Fatalf("should copy just one file")
									}

									if files[0].Content != bar {
										t.Fatalf("expected content 'bar', got '%s'", files[0].Content)
									}

									called = true

									return nil
								},
							},
						},
					},
				},
			},
		},
		currentState: containersState{
			f: &hostConfiguredContainer{
				configFiles: map[string]string{},
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						config: types.ContainerConfig{
							Image: f,
						},
						runtimeConfig: &runtime.FakeConfig{
							Runtime: &runtime.Fake{
								CreateF: func(config *types.ContainerConfig) (string, error) {
									return f, nil
								},
								StatusF: func(id string) (types.ContainerStatus, error) {
									return types.ContainerStatus{}, nil
								},
							},
						},
					},
				},
			},
		},
	}

	if err := c.ensureConfigured(f); err != nil {
		t.Fatalf("Ensure configured should succeed, got: %v", err)
	}

	if !called {
		t.Fatalf("should call Copy on container")
	}
}

func TestEnsureConfiguredFreshState(t *testing.T) {
	called := false

	f := foo
	cf := map[string]string{
		f: bar,
	}

	c := &containers{
		desiredState: containersState{
			f: &hostConfiguredContainer{
				configFiles: cf,
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						config: types.ContainerConfig{
							Image: f,
						},
						runtimeConfig: &runtime.FakeConfig{
							Runtime: &runtime.Fake{
								CreateF: func(config *types.ContainerConfig) (string, error) {
									return f, nil
								},
								StatusF: func(id string) (types.ContainerStatus, error) {
									return types.ContainerStatus{}, nil
								},
								CopyF: func(id string, files []*types.File) error {
									if id != f {
										t.Errorf("should copy to configuration container '%s', not to '%s'", f, id)
									}

									if len(files) != len(cf) {
										t.Fatalf("should copy just one file")
									}

									if files[0].Content != bar {
										t.Fatalf("expected content 'bar', got '%s'", files[0].Content)
									}

									called = true

									return nil
								},
							},
						},
					},
				},
			},
		},
		currentState: containersState{},
	}

	if err := c.ensureConfigured(f); err != nil {
		t.Fatalf("Ensure configured should succeed, got: %v", err)
	}

	if !called {
		t.Fatalf("should call Copy on container")
	}
}

// selectRuntime()
func TestSelectRuntime(t *testing.T) {
	c := &container{
		base: base{
			runtimeConfig: &docker.Config{},
		},
	}

	if err := c.selectRuntime(); err != nil {
		t.Fatalf("selecting runtime should succeed, got: %v", err)
	}

	if c.base.runtime == nil {
		t.Fatalf("selectRuntime should set container runtime")
	}
}

// DesiredState
func TestContainersDesiredStateEmpty(t *testing.T) {
	c := &containers{}

	e := ContainersState{}

	if diff := cmp.Diff(e, c.DesiredState()); diff != "" {
		t.Fatalf("Unexpected diff: %s", diff)
	}
}

func TestContainersDesiredStateOldID(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{
							Image: "a",
						},
						runtimeConfig: docker.DefaultConfig(),
					},
				},
			},
		},
		previousState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{
							Image: "b",
						},
						runtimeConfig: docker.DefaultConfig(),
						status: types.ContainerStatus{
							Status: "running",
							ID:     "foo",
						},
					},
				},
			},
		},
	}

	e := ContainersState{
		foo: {
			Container: Container{
				Config: types.ContainerConfig{Image: "a"},
				Status: types.ContainerStatus{ID: "foo", Status: "running"},
				Runtime: RuntimeConfig{
					Docker: docker.DefaultConfig(),
				},
			},
			ConfigFiles: map[string]string{},
		},
	}

	if diff := cmp.Diff(e, c.DesiredState()); diff != "" {
		t.Fatalf("Unexpected diff: %s", diff)
	}
}

func TestContainersDesiredStateStatusRunning(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{
							Image: "a",
						},
						runtimeConfig: docker.DefaultConfig(),
					},
				},
			},
		},
		previousState: containersState{
			foo: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{
							Image: "a",
						},
						runtimeConfig: docker.DefaultConfig(),
						status: types.ContainerStatus{
							Status: StatusMissing,
						},
					},
				},
			},
		},
	}

	e := ContainersState{
		foo: {
			Container: Container{
				Config: types.ContainerConfig{Image: "a"},
				Status: types.ContainerStatus{Status: "running"},
				Runtime: RuntimeConfig{
					Docker: docker.DefaultConfig(),
				},
			},
			ConfigFiles: map[string]string{},
		},
	}

	if diff := cmp.Diff(e, c.DesiredState()); diff != "" {
		t.Fatalf("Unexpected diff: %s", diff)
	}
}
