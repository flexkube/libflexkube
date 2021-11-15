package container

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/container/runtime"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

// New() tests.
func TestContainersNew(t *testing.T) {
	t.Parallel()

	GetContainers(t)
}

func GetContainers(t *testing.T) ContainersInterface {
	t.Helper()

	containersConfig := &Containers{
		DesiredState: ContainersState{
			testContainerName: &HostConfiguredContainer{
				Host: host.Host{
					DirectConfig: &direct.Config{},
				},
				Container: Container{
					Runtime: RuntimeConfig{
						Docker: &docker.Config{},
					},
					Config: types.ContainerConfig{
						Name:  testConfigContainerName,
						Image: "busybox:latest",
					},
				},
			},
		},
	}

	c, err := containersConfig.New()
	if err != nil {
		t.Fatalf("Creating empty containers object should work, got: %v", err)
	}

	return c
}

// Containers() tests.
func TestContainersContainers(t *testing.T) {
	t.Parallel()

	testContainers := &containers{}

	if !reflect.DeepEqual(testContainers, testContainers.Containers()) {
		t.Fatalf("Containers() should return self")
	}
}

// CheckCurrentState() tests.
func TestContainersCheckCurrentStateNew(t *testing.T) {
	t.Parallel()

	c := GetContainers(t)

	if err := c.CheckCurrentState(); err != nil {
		t.Fatalf("Checking current state for new Containers should work, got: %v", err)
	}
}

// CurrentStateToYaml() tests.
func TestContainersCurrentStateToYAML(t *testing.T) {
	t.Parallel()

	c := GetContainers(t)

	if _, err := c.StateToYaml(); err != nil {
		t.Fatalf("Getting current state in YAML format should work, got: %v", err)
	}
}

// ToExported() tests.
func TestContainersToExported(t *testing.T) {
	t.Parallel()

	c := GetContainers(t)

	c.ToExported()
}

// FromYaml() tests.
func TestContainersFromYamlBad(t *testing.T) {
	t.Parallel()

	if _, err := FromYaml([]byte{}); err == nil {
		t.Fatalf("Creating containers from empty YAML should fail")
	}
}

func TestContainersFromYaml(t *testing.T) {
	t.Parallel()

	containersConfigRaw := `
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

	if _, err := FromYaml([]byte(containersConfigRaw)); err != nil {
		t.Fatalf("Creating containers from valid YAML should work, got: %v", err)
	}
}

// filesToUpdate() tests.
func TestFilesToUpdateEmpty(t *testing.T) {
	t.Parallel()

	expected := []string{testConfigPath}

	hcc := hostConfiguredContainer{
		configFiles: map[string]string{
			testConfigPath: testConfigContent,
		},
	}

	if v := filesToUpdate(hcc, nil); !reflect.DeepEqual(expected, v) {
		t.Fatalf("Expected %v, got %v", expected, hcc)
	}
}

// Validate() tests.
func TestValidateEmpty(t *testing.T) {
	t.Parallel()

	cc := &Containers{}

	if err := cc.Validate(); err == nil {
		t.Fatalf("Empty containers object shouldn't be valid")
	}
}

func TestValidateNil(t *testing.T) {
	t.Parallel()

	var cc Containers

	if err := cc.Validate(); err == nil {
		t.Fatalf("Nil containers object shouldn't be valid")
	}
}

func TestValidateNoContainers(t *testing.T) {
	t.Parallel()

	cc := &Containers{
		DesiredState:  ContainersState{},
		PreviousState: ContainersState{},
	}

	if err := cc.Validate(); err == nil {
		t.Fatalf("Containers object without any containers shouldn't be valid")
	}
}

func TestValidateBadDesiredContainers(t *testing.T) {
	t.Parallel()

	containersConfig := &Containers{
		DesiredState: ContainersState{
			testContainerName: &HostConfiguredContainer{},
		},
		PreviousState: ContainersState{},
	}

	if err := containersConfig.Validate(); err == nil {
		t.Fatalf("Containers object with bad desired container shouldn't be valid")
	}
}

func TestValidateBadCurrentContainers(t *testing.T) {
	t.Parallel()

	containersConfig := &Containers{
		DesiredState: ContainersState{},
		PreviousState: ContainersState{
			testContainerName: &HostConfiguredContainer{},
		},
	}

	if err := containersConfig.Validate(); err == nil {
		t.Fatalf("Containers object with bad current container shouldn't be valid")
	}
}

// isUpdatable() tests.
func TestIsUpdatableWithoutCurrentState(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{},
		},
	}

	if err := testContainers.isUpdatable(testContainerName); err == nil {
		t.Fatalf("Container without current state shouldn't be updatable.")
	}
}

func TestIsUpdatableToBeDeleted(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{},
		},
	}

	if err := testContainers.isUpdatable(testContainerName); err == nil {
		t.Fatalf("Container without desired state shouldn't be updatable.")
	}
}

func TestIsUpdatable(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{},
		},
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{},
		},
	}

	if err := testContainers.isUpdatable(testContainerName); err != nil {
		t.Fatalf("Container with current and desired state should be updatable, got: %v", err)
	}
}

// diffHost() tests.
func TestDiffHostNotUpdatable(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{},
		},
	}

	if _, err := testContainers.diffHost(testContainerName); err == nil {
		t.Fatalf("Not updatable container shouldn't return diff")
	}
}

func TestDiffHostNoDiff(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				host: host.Host{},
			},
		},
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{
				host: host.Host{},
			},
		},
	}

	diff, err := testContainers.diffHost(testContainerName)
	if err != nil {
		t.Fatalf("Updatable container should return diff, got: %v", err)
	}

	if diff != "" {
		t.Fatalf("Container without host updates shouldn't return diff, got: %s", diff)
	}
}

func TestDiffHost(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
			},
		},
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{
				host: host.Host{},
			},
		},
	}

	diff, err := testContainers.diffHost(testContainerName)
	if err != nil {
		t.Fatalf("Updatable container should return diff, got: %v", err)
	}

	if diff == "" {
		t.Fatalf("Container with host updates should return diff")
	}
}

// diffContainer() tests.
func TestDiffContainerNotUpdatable(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{},
		},
	}

	if _, err := testContainers.diffContainer(testContainerName); err == nil {
		t.Fatalf("Not updatable container shouldn't return diff")
	}
}

func TestDiffContainerNoDiff(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
			},
		},
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
			},
		},
	}

	diff, err := testContainers.diffContainer(testContainerName)
	if err != nil {
		t.Fatalf("Updatable container should return diff, got: %v", err)
	}

	if diff != "" {
		t.Fatalf("Container without host updates shouldn't return diff, got: %s", diff)
	}
}

func TestDiffContainer(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
			},
		},
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{
							Name: testConfigContainerName,
						},
					},
				},
			},
		},
	}

	diff, err := testContainers.diffContainer(testContainerName)
	if err != nil {
		t.Fatalf("Updatable container should return diff, got: %v", err)
	}

	if diff == "" {
		t.Fatalf("Container with config updates should return diff")
	}
}

func TestDiffContainerRuntimeConfig(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				container: &container{
					base: base{
						config:        types.ContainerConfig{},
						runtimeConfig: &docker.Config{},
					},
				},
			},
		},
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
						runtimeConfig: &docker.Config{
							Host: "foo",
						},
					},
				},
			},
		},
	}

	diff, err := testContainers.diffContainer(testContainerName)
	if err != nil {
		t.Fatalf("Updatable container should return diff, got: %v", err)
	}

	if diff == "" {
		t.Fatalf("Container with runtime config updates should return diff")
	}
}

// ensureRunning() tests.
func TestEnsureRunningNonExistent(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		currentState: containersState{},
	}

	if err := ensureRunning(testContainers.currentState[testAnotherContainerName]); err == nil {
		t.Fatalf("Ensuring that non existing container is running should fail")
	}
}

func TestEnsureRunning(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{
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

	if err := ensureRunning(testContainers.currentState[testContainerName]); err != nil {
		t.Fatalf("Ensuring that running container is running should succeed, got: %v", err)
	}
}

// ensureExists() tests.
func TestEnsureExistsAlreadyExists(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{
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

	if err := testContainers.ensureExists(testContainerName); err != nil {
		t.Fatalf("Ensuring that existing container exists should succeed, got: %v", err)
	}
}

func TestEnsureExistsFailCreate(t *testing.T) {
	t.Parallel()

	failingCreateRuntime := fakeRuntime()
	failingCreateRuntime.CreateF = func(config *types.ContainerConfig) (string, error) {
		return "", fmt.Errorf("create fail")
	}

	testContainers := &containers{
		currentState: containersState{},
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						config:        types.ContainerConfig{},
						runtimeConfig: asRuntime(failingCreateRuntime),
					},
				},
			},
		},
	}

	if err := testContainers.ensureExists(testContainerName); err == nil {
		t.Fatalf("Ensuring that new container exists should propagate create error")
	}

	if len(testContainers.currentState) != 0 {
		t.Fatalf("If creation failed, current state should not be updated")
	}
}

func TestEnsureExistsFailStart(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		currentState: containersState{},
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				hooks: &Hooks{},
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						config:        types.ContainerConfig{},
						runtimeConfig: asRuntime(failingStartRuntime()),
					},
				},
			},
		},
	}

	if err := testContainers.ensureExists(testContainerName); err == nil {
		t.Fatalf("Ensuring that new container exists should fail")
	}

	if _, ok := testContainers.currentState[testContainerName]; !ok {
		t.Fatalf("EnsureExists should save state of created container even if starting failed")
	}
}

func TestEnsureExist(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		currentState: containersState{},
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				hooks: &Hooks{},
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						config:        types.ContainerConfig{},
						runtimeConfig: asRuntime(fakeRuntime()),
					},
				},
			},
		},
	}

	if err := testContainers.ensureExists(testContainerName); err != nil {
		t.Fatalf("Ensuring that new container exists should succeed, got: %v", err)
	}

	if _, ok := testContainers.currentState[testContainerName]; !ok {
		t.Fatalf("EnsureExists should save state of created container when creation succeeds")
	}
}

// ensureHost() tests.
func TestEnsureHostNoDiff(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
			},
		},
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
			},
		},
	}

	if err := testContainers.ensureHost(testContainerName); err != nil {
		t.Fatalf("Ensuring that container's host configuration is up to date should succeed, got: %v", err)
	}
}

func TestEnsureHostFailStart(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						status: types.ContainerStatus{
							ID: testContainerID,
						},
						runtimeConfig: asRuntime(fakeRuntime()),
					},
				},
			},
		},
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				hooks: &Hooks{},
				host: host.Host{
					DirectConfig: &direct.Config{
						Dummy: "foo",
					},
				},
				container: &container{
					base: base{
						runtimeConfig: asRuntime(failingStartRuntime()),
					},
				},
			},
		},
	}

	if err := testContainers.ensureHost(testContainerName); err == nil {
		t.Fatalf("Ensuring that container's host configuration is up to date should fail")
	}

	if gotID := testContainers.currentState[testContainerName].container.Status().ID; gotID != testAnotherContainerID {
		t.Logf("Expected container ID %q, got %q", testAnotherContainerID, gotID)

		t.Fatalf("Ensure host should persist state changes even if process fails")
	}
}

func TestEnsureHost(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				hooks: &Hooks{},
				host: host.Host{
					DirectConfig: &direct.Config{
						Dummy: "foo",
					},
				},
				container: &container{
					base: base{
						runtimeConfig: asRuntime(fakeRuntime()),
					},
				},
			},
		},
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						status: types.ContainerStatus{
							ID: testContainerID,
						},
						runtimeConfig: asRuntime(fakeRuntime()),
					},
				},
			},
		},
	}

	if err := testContainers.ensureHost(testContainerName); err != nil {
		t.Fatalf("Ensuring that container's host configuration is up to date should succeed, got: %v", err)
	}

	if testContainers.currentState[testContainerName].container.Status().ID != testAnotherContainerID {
		t.Fatalf("Ensure host should persist state changes even if process fails")
	}
}

// ensureContainer() tests.
func TestEnsureContainerNoDiff(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
			},
		},
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
			},
		},
	}

	if err := testContainers.ensureContainer(testContainerName); err != nil {
		t.Fatalf("Ensuring that container configuration is up to date should succeed, got: %v", err)
	}
}

func TestEnsureContainerFailStart(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						status: types.ContainerStatus{
							ID: testContainerID,
						},
						config: types.ContainerConfig{
							Image: testImage,
						},
						runtimeConfig: asRuntime(fakeRuntime()),
					},
				},
			},
		},
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				hooks: &Hooks{},
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						config: types.ContainerConfig{
							Image: testImage,
						},
						runtimeConfig: asRuntime(failingStartRuntime()),
					},
				},
			},
		},
	}

	if err := testContainers.ensureContainer(testContainerName); err == nil {
		t.Fatalf("Ensuring that container configuration is up to date should fail")
	}

	if gotID := testContainers.currentState[testContainerName].container.Status().ID; gotID != testAnotherContainerID {
		t.Fatalf("Ensure container should persist state changes even if process fails")
	}
}

func TestEnsureContainer(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				hooks: &Hooks{},
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						config: types.ContainerConfig{
							Image: testImage,
						},
						runtimeConfig: asRuntime(fakeRuntime()),
					},
				},
			},
		},
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						status: types.ContainerStatus{
							ID: testContainerID,
						},
						config: types.ContainerConfig{
							Image: testAnotherImage,
						},
						runtimeConfig: asRuntime(fakeRuntime()),
					},
				},
			},
		},
	}

	if err := testContainers.ensureContainer(testContainerName); err != nil {
		t.Fatalf("Ensuring that container configuration is up to date should succeed, got: %v", err)
	}

	if testContainers.currentState[testContainerName].container.Status().ID != testAnotherContainerID {
		t.Fatalf("Ensure container should persist state changes even if process fails")
	}
}

// recreate() tests.
func TestRecreateNonExistent(t *testing.T) {
	t.Parallel()

	testContainers := &containers{}
	if err := testContainers.recreate(testContainerName); err == nil {
		t.Fatalf("Recreating on empty containers should fail")
	}
}

// Deploy() tests.
func TestDeployNoCurrentState(t *testing.T) {
	t.Parallel()

	testContainers := &containers{}
	if err := testContainers.Deploy(); err == nil {
		t.Fatalf("Execute without current state should fail")
	}
}

// hasUpdates() tests.
func TestHasUpdatesHost(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
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
			testContainerName: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
				host: host.Host{},
			},
		},
	}

	u, err := testContainers.hasUpdates(testContainerName)
	if err != nil {
		t.Fatalf("Checking for updates should succeed, got: %v", err)
	}

	if !u {
		t.Fatalf("Container with host changes should indicate update")
	}
}

func TestHasUpdatesConfig(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
				configFiles: map[string]string{
					testConfigPath: testConfigContent,
				},
			},
		},
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
			},
		},
	}

	u, err := testContainers.hasUpdates(testContainerName)
	if err != nil {
		t.Fatalf("Checking for updates should succeed, got: %v", err)
	}

	if !u {
		t.Fatalf("Container with configuration files changes should indicate update")
	}
}

func TestHasUpdatesContainer(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
			},
		},
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{
							Name: testConfigContainerName,
						},
					},
				},
			},
		},
	}

	u, err := testContainers.hasUpdates(testContainerName)
	if err != nil {
		t.Fatalf("Checking for updates should succeed, got: %v", err)
	}

	if !u {
		t.Fatalf("Container with container configuration changes should indicate update")
	}
}

func TestHasUpdatesNoUpdates(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
			},
		},
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{
				container: &container{
					base: base{
						config: types.ContainerConfig{},
					},
				},
			},
		},
	}

	u, err := testContainers.hasUpdates(testContainerName)
	if err != nil {
		t.Fatalf("Checking for updates should succeed, got: %v", err)
	}

	if u {
		t.Fatalf("Container with no changes should indicate no changes")
	}
}

// ensureConfigured() tests.
func TestEnsureConfiguredDisposable(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		desiredState: containersState{},
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{
				configFiles: map[string]string{},
			},
		},
	}

	if err := testContainers.ensureConfigured(testContainerName); err != nil {
		t.Fatalf("Ensure configured should succeed when container is going to be removed, got: %v", err)
	}
}

func TestEnsureConfiguredNoUpdates(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				configFiles: map[string]string{},
			},
		},
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{
				configFiles: map[string]string{},
			},
		},
	}

	if err := testContainers.ensureConfigured(testContainerName); err != nil {
		t.Fatalf("Ensure configured should succeed, got: %v", err)
	}
}

const (
	testImage                = "test-image"
	testAnotherImage         = "test-another-image"
	testConfigContent        = "test-config-content"
	testConfigPath           = "/tmp/foo"
	testContainerID          = "test-container-id"
	testAnotherContainerID   = "test-another-container-id"
	testContainerName        = "test-container-name"
	testConfigContainerName  = "test-config-container-name"
	testAnotherContainerName = "test-another-container-name"
)

func TestEnsureConfigured(t *testing.T) {
	t.Parallel()

	called := false

	testConfigFiles := map[string]string{
		testConfigPath: testConfigContent,
	}

	testContainers := &containers{
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{
				configFiles: map[string]string{},
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						config: types.ContainerConfig{
							Image: testImage,
						},
						runtimeConfig: asRuntime(fakeRuntime()),
					},
				},
			},
		},
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				configFiles: testConfigFiles,
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						config: types.ContainerConfig{
							Image: testImage,
						},
						runtimeConfig: asRuntime(testCopyingRuntime(t, &called, testContainerID, testConfigFiles)),
					},
				},
			},
		},
	}

	if err := testContainers.ensureConfigured(testContainerName); err != nil {
		t.Fatalf("Ensure configured should succeed, got: %v", err)
	}

	if !called {
		t.Fatalf("Should call Copy on container")
	}
}

func TestEnsureConfiguredFreshState(t *testing.T) {
	t.Parallel()

	called := false

	testConfigFiles := map[string]string{
		testConfigPath: testConfigContent,
	}

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				configFiles: testConfigFiles,
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						config: types.ContainerConfig{
							Image: testImage,
						},
						runtimeConfig: asRuntime(testCopyingRuntime(t, &called, testContainerID, testConfigFiles)),
					},
				},
			},
		},
		currentState: containersState{},
	}

	if err := testContainers.ensureConfigured(testContainerName); err != nil {
		t.Fatalf("Ensure configured should succeed, got: %v", err)
	}

	if !called {
		t.Fatalf("Should call Copy on container")
	}
}

func TestEnsureConfiguredNoStateUpdateOnFail(t *testing.T) {
	t.Parallel()

	testConfigFiles := map[string]string{
		testConfigPath: testConfigContent,
	}

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
				configFiles: testConfigFiles,
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						config: types.ContainerConfig{
							Image: testImage,
						},
						runtimeConfig: &runtime.FakeConfig{
							Runtime: &runtime.Fake{
								CreateF: func(config *types.ContainerConfig) (string, error) {
									return testContainerID, nil
								},
								StatusF: func(id string) (types.ContainerStatus, error) {
									return types.ContainerStatus{}, nil
								},
								CopyF: func(id string, files []*types.File) error {
									return fmt.Errorf("fail")
								},
							},
						},
					},
				},
			},
		},
		currentState: containersState{},
	}

	if err := testContainers.ensureConfigured(testContainerName); err == nil {
		t.Fatalf("Ensure configured should fail")
	}

	if len(testContainers.currentState) != 0 {
		t.Fatalf("If no files has been updated, current state should not be set")
	}
}

// selectRuntime() tests.
func TestSelectRuntime(t *testing.T) {
	t.Parallel()

	testContainer := &container{
		base: base{
			runtimeConfig: &docker.Config{},
		},
	}

	if err := testContainer.selectRuntime(); err != nil {
		t.Fatalf("Selecting runtime should succeed, got: %v", err)
	}

	if testContainer.base.runtime == nil {
		t.Fatalf("SelectRuntime should set container runtime")
	}
}

// DesiredState() tests.
func TestContainersDesiredStateEmpty(t *testing.T) {
	t.Parallel()

	testContainers := &containers{}

	e := ContainersState{}

	if diff := cmp.Diff(e, testContainers.DesiredState()); diff != "" {
		t.Fatalf("Unexpected diff: %s", diff)
	}
}

func TestContainersDesiredStateOldID(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
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
			testContainerName: &hostConfiguredContainer{
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

	expectedContainersState := ContainersState{
		testContainerName: {
			Container: Container{
				Config: types.ContainerConfig{Image: "a"},
				Status: &types.ContainerStatus{ID: "foo", Status: "running"},
				Runtime: RuntimeConfig{
					Docker: docker.DefaultConfig(),
				},
			},
			ConfigFiles: map[string]string{},
		},
	}

	if diff := cmp.Diff(testContainers.DesiredState(), expectedContainersState); diff != "" {
		t.Fatalf("Unexpected diff: %s", diff)
	}
}

func TestContainersDesiredStateStatusRunning(t *testing.T) {
	t.Parallel()

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hostConfiguredContainer{
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
			testContainerName: &hostConfiguredContainer{
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

	expectedContainersState := ContainersState{
		testContainerName: {
			Container: Container{
				Config: types.ContainerConfig{Image: "a"},
				Status: &types.ContainerStatus{Status: "running"},
				Runtime: RuntimeConfig{
					Docker: docker.DefaultConfig(),
				},
			},
			ConfigFiles: map[string]string{},
		},
	}

	if diff := cmp.Diff(testContainers.DesiredState(), expectedContainersState); diff != "" {
		t.Fatalf("Unexpected diff: %s", diff)
	}
}

// updateExistingContainers() tests.
func TestUpdateExistingContainersRemoveAllOld(t *testing.T) {
	t.Parallel()

	fooRuntime := fakeRuntime()
	fooRuntime.StatusF = func(id string) (types.ContainerStatus, error) {
		return types.ContainerStatus{
			Status: "running",
			ID:     "foo",
		}, nil
	}

	barRuntime := fakeRuntime()
	barRuntime.StatusF = func(id string) (types.ContainerStatus, error) {
		return types.ContainerStatus{
			Status: "running",
			ID:     "bar",
		}, nil
	}

	testContainers := &containers{
		desiredState: containersState{},
		currentState: containersState{
			testContainerName: &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						status: types.ContainerStatus{
							Status: "running",
							ID:     "foo",
						},
						runtimeConfig: asRuntime(fooRuntime),
					},
				},
			},
			testAnotherContainerName: &hostConfiguredContainer{
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
				container: &container{
					base: base{
						status: types.ContainerStatus{
							Status: "running",
							ID:     "bar",
						},
						runtimeConfig: asRuntime(barRuntime),
					},
				},
			},
		},
	}

	if err := testContainers.updateExistingContainers(); err != nil {
		t.Fatalf("Updating existing containers should succeed, got: %v", err)
	}

	if len(testContainers.desiredState) != len(testContainers.currentState) {
		t.Fatalf("All containers from current state should be removed")
	}
}

// ensureCurrentContainer() tests.
func TestEnsureCurrentContainer(t *testing.T) {
	t.Parallel()

	hcc := hostConfiguredContainer{
		hooks: &Hooks{},
		host: host.Host{
			DirectConfig: &direct.Config{},
		},
		container: &container{
			base: base{
				status: types.ContainerStatus{
					Status: "stopped",
					ID:     "foo",
				},
				runtimeConfig: &docker.Config{
					Host: "unix:///nonexistent",
				},
			},
		},
	}

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hcc,
		},
		currentState: containersState{
			testContainerName: &hcc,
		},
	}

	_, err := testContainers.ensureCurrentContainer(testContainerName, hcc)
	if err == nil {
		t.Fatalf("Ensure stopped container should try to start the container and fail")
	}

	if !strings.Contains(err.Error(), "Is the docker daemon running?") {
		t.Fatalf("Ensuring stopped container should fail to contact non existing runtime, got: %v", err)
	}
}

func TestEnsureCurrentContainerNonExisting(t *testing.T) {
	t.Parallel()

	hcc := hostConfiguredContainer{
		hooks: &Hooks{},
		host: host.Host{
			DirectConfig: &direct.Config{},
		},
		container: &container{
			base: base{
				status: types.ContainerStatus{
					Status: "",
					ID:     "",
				},
				runtimeConfig: &docker.Config{
					Host: "unix:///nonexistent",
				},
			},
		},
	}

	testContainers := &containers{
		desiredState: containersState{
			testContainerName: &hcc,
		},
		currentState: containersState{
			testContainerName: &hcc,
		},
	}

	if _, err := testContainers.ensureCurrentContainer(testContainerName, hcc); err != nil {
		t.Fatalf("Ensure stopped container should not fail on non-existing container, got: %v", err)
	}

	if len(testContainers.currentState) > 0 {
		t.Fatalf("Ensuring removed container should remove it from current state to trigger creation")
	}
}

func fakeRuntime() *runtime.Fake {
	return &runtime.Fake{
		CreateF: func(config *types.ContainerConfig) (string, error) {
			return testContainerID, nil
		},
		StatusF: func(id string) (types.ContainerStatus, error) {
			return types.ContainerStatus{
				ID: testAnotherContainerID,
			}, nil
		},
		DeleteF: func(id string) error {
			return nil
		},
		StartF: func(id string) error {
			return nil
		},
		StopF: func(id string) error {
			return nil
		},
	}
}

func failingStartRuntime() *runtime.Fake {
	r := fakeRuntime()
	r.StartF = func(string) error {
		return fmt.Errorf("starting")
	}

	return r
}

func testCopyingRuntime(t *testing.T, called *bool, containerID string, config map[string]string) *runtime.Fake {
	t.Helper()

	testRuntime := fakeRuntime()
	testRuntime.CopyF = func(id string, files []*types.File) error {
		if id != containerID {
			t.Errorf("Should copy to configuration container %q, not to %q", containerID, id)
		}

		if len(files) != len(config) {
			t.Fatalf("Should copy just one file")
		}

		if files[0].Content != testConfigContent {
			t.Fatalf("Expected content %q, got %q", testConfigContent, files[0].Content)
		}

		*called = true

		return nil
	}

	return testRuntime
}

func asRuntime(r *runtime.Fake) *runtime.FakeConfig {
	return &runtime.FakeConfig{
		Runtime: r,
	}
}
