package container

import (
	"reflect"
	"testing"

	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

// New()
func TestContainersNew(t *testing.T) {
	GetContainers(t)
}

func GetContainers(t *testing.T) ContainersInterface {
	cc := &Containers{
		DesiredState: ContainersState{
			"foo": &HostConfiguredContainer{
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
			},
		},
	}

	c, err := cc.New()
	if err != nil {
		t.Fatalf("Creating empty containers object should work, got: %v", err)
	}

	return c
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

	_, err := c.CurrentStateToYaml()
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

func TestFilesToUpdateEmpty(t *testing.T) {
	expected := []string{"foo"}

	d := hostConfiguredContainer{
		configFiles: map[string]string{
			"foo": "bar",
		},
	}

	if v := filesToUpdate(d, nil); !reflect.DeepEqual(expected, v) {
		t.Fatalf("Expected %v, got %v", expected, d)
	}
}

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
			"foo": &HostConfiguredContainer{},
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
			"foo": &HostConfiguredContainer{},
		},
	}

	if err := cc.Validate(); err == nil {
		t.Fatalf("Containers object with bad current container shouldn't be valid")
	}
}

func TestIsUpdatableWithoutCurrentState(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			"foo": &hostConfiguredContainer{},
		},
	}

	if err := c.isUpdatable("foo"); err == nil {
		t.Fatalf("Container without current state shouldn't be updatable.")
	}
}

func TestIsUpdatableToBeDeleted(t *testing.T) {
	c := &containers{
		currentState: containersState{
			"foo": &hostConfiguredContainer{},
		},
	}

	if err := c.isUpdatable("foo"); err == nil {
		t.Fatalf("Container without desired state shouldn't be updatable.")
	}
}

func TestIsUpdatable(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			"foo": &hostConfiguredContainer{},
		},
		currentState: containersState{
			"foo": &hostConfiguredContainer{},
		},
	}

	if err := c.isUpdatable("foo"); err != nil {
		t.Fatalf("Container with current and desired state should be updatable, got: %v", err)
	}
}

func TestDiffHostNotUpdatable(t *testing.T) {
	c := &containers{
		currentState: containersState{
			"foo": &hostConfiguredContainer{},
		},
	}

	if _, err := c.diffHost("foo"); err == nil {
		t.Fatalf("Not updatable container shouldn't return diff")
	}
}

func TestDiffHostNoDiff(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			"foo": &hostConfiguredContainer{
				host: host.Host{},
			},
		},
		currentState: containersState{
			"foo": &hostConfiguredContainer{
				host: host.Host{},
			},
		},
	}

	diff, err := c.diffHost("foo")
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
			"foo": &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
			},
		},
		currentState: containersState{
			"foo": &hostConfiguredContainer{
				host: host.Host{},
			},
		},
	}

	diff, err := c.diffHost("foo")
	if err != nil {
		t.Fatalf("Updatable container should return diff, got: %v", err)
	}

	if diff == "" {
		t.Fatalf("Container with host updates should return diff")
	}
}

func TestDiffContainerNotUpdatable(t *testing.T) {
	c := &containers{
		currentState: containersState{
			"foo": &hostConfiguredContainer{},
		},
	}

	if _, err := c.diffContainer("foo"); err == nil {
		t.Fatalf("Not updatable container shouldn't return diff")
	}
}

func TestDiffContainerNoDiff(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			"foo": &hostConfiguredContainer{
				container: Container{
					Config: types.ContainerConfig{},
				},
			},
		},
		currentState: containersState{
			"foo": &hostConfiguredContainer{
				container: Container{
					Config: types.ContainerConfig{},
				},
			},
		},
	}

	diff, err := c.diffContainer("foo")
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
			"foo": &hostConfiguredContainer{
				container: Container{
					Config: types.ContainerConfig{},
				},
			},
		},
		currentState: containersState{
			"foo": &hostConfiguredContainer{
				container: Container{
					Config: types.ContainerConfig{
						Name: "foo",
					},
				},
			},
		},
	}

	diff, err := c.diffContainer("foo")
	if err != nil {
		t.Fatalf("Updatable container should return diff, got: %v", err)
	}

	if diff == "" {
		t.Fatalf("Container with config updates should return diff")
	}
}

func TestEnsureRunningNonExistent(t *testing.T) {
	c := &containers{
		currentState: containersState{},
	}

	if err := ensureRunning(c.currentState["bar"]); err == nil {
		t.Fatalf("Ensuring that non existing container is running should fail")
	}
}

func TestEnsureRunning(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			"foo": &hostConfiguredContainer{
				container: Container{
					Config: types.ContainerConfig{},
				},
			},
		},
		currentState: containersState{
			"foo": &hostConfiguredContainer{
				container: Container{
					Config: types.ContainerConfig{},
					Status: types.ContainerStatus{
						ID:     "existing",
						Status: "running",
					},
				},
			},
		},
	}

	if err := ensureRunning(c.currentState["foo"]); err != nil {
		t.Fatalf("Ensuring that running container is running should succeed, got: %v", err)
	}
}

func TestEnsureExists(t *testing.T) {
	c := &containers{
		currentState: containersState{
			"foo": &hostConfiguredContainer{
				container: Container{
					Config: types.ContainerConfig{},
					Status: types.ContainerStatus{
						ID: "existing",
					},
				},
			},
		},
	}

	if err := c.ensureExists("foo"); err != nil {
		t.Fatalf("Ensuring that existing container exists should succeed, got: %v", err)
	}
}

func TestEnsureHostNoDiff(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			"foo": &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
			},
		},
		currentState: containersState{
			"foo": &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
			},
		},
	}

	if err := c.ensureHost("foo"); err != nil {
		t.Fatalf("Ensuring that container's host configuration is up to date should succeed, got: %v", err)
	}
}

func TestEnsureContainerNoDiff(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			"foo": &hostConfiguredContainer{
				container: Container{
					Config: types.ContainerConfig{},
				},
			},
		},
		currentState: containersState{
			"foo": &hostConfiguredContainer{
				container: Container{
					Config: types.ContainerConfig{},
				},
			},
		},
	}

	if err := c.ensureContainer("foo"); err != nil {
		t.Fatalf("Ensuring that container configuration is up to date should succeed, got: %v", err)
	}
}

// recreate()
func TestRecreateNonExistent(t *testing.T) {
	c := &containers{}
	if err := c.recreate("foo"); err == nil {
		t.Fatalf("Recreating on empty containers should fail")
	}
}

// Execute()
func TestExecuteNoCurrentState(t *testing.T) {
	c := &containers{}
	if err := c.Execute(); err == nil {
		t.Fatalf("Execute without current state should fail")
	}
}

// hasUpdates()
func TestHasUpdatesHost(t *testing.T) {
	c := &containers{
		desiredState: containersState{
			"foo": &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
			},
		},
		currentState: containersState{
			"foo": &hostConfiguredContainer{
				host: host.Host{},
			},
		},
	}

	u, err := c.hasUpdates("foo")
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
			"foo": &hostConfiguredContainer{
				configFiles: map[string]string{
					"foo": "foo",
				},
			},
		},
		currentState: containersState{
			"foo": &hostConfiguredContainer{},
		},
	}

	u, err := c.hasUpdates("foo")
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
			"foo": &hostConfiguredContainer{
				container: Container{
					Config: types.ContainerConfig{},
				},
			},
		},
		currentState: containersState{
			"foo": &hostConfiguredContainer{
				container: Container{
					Config: types.ContainerConfig{
						Name: "foo",
					},
				},
			},
		},
	}

	u, err := c.hasUpdates("foo")
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
			"foo": &hostConfiguredContainer{},
		},
		currentState: containersState{
			"foo": &hostConfiguredContainer{},
		},
	}

	u, err := c.hasUpdates("foo")
	if err != nil {
		t.Fatalf("Checking for updates should succeed, got: %v", err)
	}

	if u {
		t.Fatalf("Container with no changes should indicate no changes")
	}
}
