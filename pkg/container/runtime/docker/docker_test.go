package docker_test

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"testing"

	dockertypes "github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	networktypes "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/google/go-cmp/cmp"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
)

// New() tests.
func TestNewClient(t *testing.T) {
	t.Parallel()

	// TODO does this kind of simple tests make sense? Integration tests calls the same thing
	// anyway. Or maybe we should simply skip error checking in itegration tests to simplify them?
	c := &docker.Config{
		ClientGetter: func(...client.Opt) (docker.Client, error) { return nil, nil },
	}

	if _, err := c.New(); err != nil {
		t.Fatalf("Creating new docker client should work, got: %s", err)
	}
}

// getClient() tests.
func TestNewClientWithHost(t *testing.T) {
	t.Parallel()

	config := &docker.Config{
		Host: "unix:///foo.sock",
		ClientGetter: func(opts ...client.Opt) (docker.Client, error) {
			for k, v := range opts {
				/*
					if dh := c.DaemonHost(); dh != config.Host {
						t.Fatalf("Client with host set should have %q as host, got: %q", config.Host, dh)
					}*/
				t.Log(k, v)
			}

			return nil, nil
		},
	}

	if _, err := config.New(); err != nil {
		t.Fatalf("Creating new docker client should work, got: %s", err)
	}
}

// sanitizeImageName() tests.
func TestSanitizeImageName(t *testing.T) {
	t.Parallel()

	expectedImage := "foo:latest"

	testConfig := &docker.Config{
		ClientGetter: func(...client.Opt) (docker.Client, error) {
			return &docker.FakeClient{
				ContainerCreateF: func(
					ctx context.Context,
					config *containertypes.Config,
					hostConfig *containertypes.HostConfig,
					networkingConfig *networktypes.NetworkingConfig,
					platform *v1.Platform,
					containerName string,
				) (containertypes.ContainerCreateCreatedBody, error) {
					return containertypes.ContainerCreateCreatedBody{}, nil
				},
				ImageListF: func(
					ctx context.Context,
					options dockertypes.ImageListOptions,
				) ([]dockertypes.ImageSummary, error) {
					return []dockertypes.ImageSummary{
						{
							ID: "nonemptystring",
							RepoTags: []string{
								expectedImage,
							},
						},
					}, nil
				},
				ImagePullF: func(
					ctx context.Context,
					ref string,
					options dockertypes.ImagePullOptions,
				) (io.ReadCloser, error) {
					t.Fatalf("Unexpected call to image pull")

					return nil, nil
				},
			}, nil
		},
	}

	testClient, err := testConfig.New()
	if err != nil {
		t.Fatalf("Unexpected error creating test client: %v", err)
	}

	containerConfig := &types.ContainerConfig{
		Image: "foo",
	}

	if _, err := testClient.Create(containerConfig); err != nil {
		t.Fatalf("Unexpected error creating test container: %v", err)
	}
}

func TestSanitizeImageNameWithTag(t *testing.T) {
	t.Parallel()

	testConfig := &docker.Config{
		ClientGetter: func(...client.Opt) (docker.Client, error) {
			return &docker.FakeClient{
				ContainerCreateF: func(
					ctx context.Context,
					config *containertypes.Config,
					hostConfig *containertypes.HostConfig,
					networkingConfig *networktypes.NetworkingConfig,
					platform *v1.Platform,
					containerName string,
				) (containertypes.ContainerCreateCreatedBody, error) {
					if config.Image != "foo:v0.1.0" {
						t.Fatal()
					}

					return containertypes.ContainerCreateCreatedBody{}, nil
				},
			}, nil
		},
	}

	testClient, err := testConfig.New()
	if err != nil {
		t.Fatalf("Unexpected error creating test client: %v", err)
	}

	containerConfig := &types.ContainerConfig{
		Image: "foo:v0.1.0",
	}

	if _, err := testClient.Create(containerConfig); err != nil {
		t.Fatalf("Unexpected error creating test container: %v", err)
	}
}

// Status() tests.
func TestStatus(t *testing.T) {
	t.Parallel()

	expectedStatus := "running"

	testConfig := &docker.Config{
		ClientGetter: func(...client.Opt) (docker.Client, error) {
			return &docker.FakeClient{
				ContainerInspectF: func(ctx context.Context, id string) (dockertypes.ContainerJSON, error) {
					return dockertypes.ContainerJSON{
						ContainerJSONBase: &dockertypes.ContainerJSONBase{
							State: &dockertypes.ContainerState{
								Status: expectedStatus,
							},
						},
					}, nil
				},
			}, nil
		},
	}

	testClient, err := testConfig.New()
	if err != nil {
		t.Fatalf("Unexpected error creating test client: %v", err)
	}

	status, err := testClient.Status("foo")
	if err != nil {
		t.Fatalf("Checking for status should succeed, got: %v", err)
	}

	if status.ID == "" {
		t.Fatalf("ID in status of existing container should not be empty")
	}

	if status.Status != expectedStatus {
		t.Fatalf("Received status should be %s, got %s", expectedStatus, status.Status)
	}
}

func TestStatusNotFound(t *testing.T) {
	t.Parallel()

	testConfig := &docker.Config{
		ClientGetter: func(...client.Opt) (docker.Client, error) {
			return &docker.FakeClient{
				ContainerInspectF: func(ctx context.Context, id string) (dockertypes.ContainerJSON, error) {
					return dockertypes.ContainerJSON{}, errdefs.NotFound(fmt.Errorf("not found"))
				},
			}, nil
		},
	}

	testClient, err := testConfig.New()
	if err != nil {
		t.Fatalf("Unexpected error creating test client: %v", err)
	}

	s, err := testClient.Status("foo")
	if err != nil {
		t.Fatalf("Checking for status should succeed, got: %v", err)
	}

	if s.ID != "" {
		t.Fatalf("ID in status of non-existing container should be empty")
	}
}

func TestStatusRuntimeError(t *testing.T) {
	t.Parallel()

	testConfig := &docker.Config{
		ClientGetter: func(...client.Opt) (docker.Client, error) {
			return &docker.FakeClient{
				ContainerInspectF: func(ctx context.Context, id string) (dockertypes.ContainerJSON, error) {
					return dockertypes.ContainerJSON{}, fmt.Errorf("can't check status of container")
				},
			}, nil
		},
	}

	testClient, err := testConfig.New()
	if err != nil {
		t.Fatalf("Unexpected error creating test client: %v", err)
	}

	if _, err := testClient.Status("foo"); err == nil {
		t.Fatalf("Checking for status should fail")
	}
}

// Copy() tests.
func TestCopyRuntimeError(t *testing.T) {
	t.Parallel()

	testConfig := &docker.Config{
		ClientGetter: func(...client.Opt) (docker.Client, error) {
			return &docker.FakeClient{
				CopyToContainerF: func(_ context.Context, _, _ string, _ io.Reader, _ dockertypes.CopyToContainerOptions) error {
					return fmt.Errorf("Copying failed")
				},
			}, nil
		},
	}

	testClient, err := testConfig.New()
	if err != nil {
		t.Fatalf("Unexpected error creating test client: %v", err)
	}

	if err := testClient.Copy("foo", []*types.File{}); err == nil {
		t.Fatalf("Should fail when runtime returns error")
	}
}

// Read() tests.
func TestReadRuntimeError(t *testing.T) {
	t.Parallel()

	testConfig := &docker.Config{
		ClientGetter: func(...client.Opt) (docker.Client, error) {
			return &docker.FakeClient{
				CopyFromContainerF: func(_ context.Context, _, path string) (io.ReadCloser, dockertypes.ContainerPathStat, error) {
					if path != defaultPath {
						t.Fatalf("Should read path %s, got %s", defaultPath, path)
					}

					return nil, dockertypes.ContainerPathStat{}, fmt.Errorf("Copying failed")
				},
			}, nil
		},
	}

	testClient, err := testConfig.New()
	if err != nil {
		t.Fatalf("Unexpected error creating test client: %v", err)
	}

	if _, err := testClient.Read("foo", []string{defaultPath}); err == nil {
		t.Fatalf("Should fail when runtime returns error")
	}
}

const (
	defaultMode = 420
	defaultPath = "/foo"
)

func TestRead(t *testing.T) {
	t.Parallel()

	testConfig := &docker.Config{
		ClientGetter: func(...client.Opt) (docker.Client, error) {
			return &docker.FakeClient{
				CopyFromContainerF: func(_ context.Context, _, _ string) (io.ReadCloser, dockertypes.ContainerPathStat, error) {
					return io.NopCloser(testTar(t)), dockertypes.ContainerPathStat{
						Name: defaultPath,
					}, nil
				},
			}, nil
		},
	}

	testClient, err := testConfig.New()
	if err != nil {
		t.Fatalf("Unexpected error creating test client: %v", err)
	}

	readFiles, err := testClient.Read("foo", []string{defaultPath})
	if err != nil {
		t.Fatalf("Reading should succeed, got: %v", err)
	}

	expectedFiles := []*types.File{
		{
			Path:    defaultPath,
			Content: "foo\n",
			Mode:    defaultMode,
			User:    "1000",
			Group:   "1000",
		},
	}

	if diff := cmp.Diff(readFiles, expectedFiles); diff != "" {
		t.Fatalf("Got unexpected files: %s", diff)
	}
}

func TestReadFileMissing(t *testing.T) {
	t.Parallel()

	testConfig := &docker.Config{
		ClientGetter: func(...client.Opt) (docker.Client, error) {
			return &docker.FakeClient{
				CopyFromContainerF: func(_ context.Context, _, _ string) (io.ReadCloser, dockertypes.ContainerPathStat, error) {
					return nil, dockertypes.ContainerPathStat{}, nil
				},
			}, nil
		},
	}

	testClient, err := testConfig.New()
	if err != nil {
		t.Fatalf("Unexpected error creating test client: %v", err)
	}

	fs, err := testClient.Read("foo", []string{defaultPath})
	if err != nil {
		t.Fatalf("Read should succeed, got: %v", err)
	}

	if len(fs) != 0 {
		t.Fatalf("Read should not return any files if the file does not exist")
	}
}

func testTar(t *testing.T) io.Reader {
	t.Helper()

	r := strings.NewReader(`H4sIAAAAAAAAA+3RQQrCMBCF4aw9RW6QTEza87SYYrA20lrx+K2g4EZs6UKE/9u8xQzMMHPM52hS
d0uHVHWmj5c8mDbVTRvvp7GOpslZbWVnhfePlDLY93zySvaF86EUCaKss+K80nbz5AXG4Vr1WqvX
DT71fav/qfm/u1/vAAAAAAAAAAAAAAAAAABYbwIOFGnRACgAAA==`)

	g, err := gzip.NewReader(base64.NewDecoder(base64.StdEncoding, r))
	if err != nil {
		t.Fatalf("Creating reader should succeed, got: %v", err)
	}

	return g
}

func TestReadVerifyTarArchive(t *testing.T) {
	t.Parallel()

	testConfig := &docker.Config{
		ClientGetter: func(...client.Opt) (docker.Client, error) {
			return &docker.FakeClient{
				CopyFromContainerF: func(_ context.Context, _, _ string) (io.ReadCloser, dockertypes.ContainerPathStat, error) {
					return io.NopCloser(strings.NewReader("asdasd")), dockertypes.ContainerPathStat{}, nil
				},
			}, nil
		},
	}

	testClient, err := testConfig.New()
	if err != nil {
		t.Fatalf("Unexpected error creating test client: %v", err)
	}

	if _, err := testClient.Read("foo", []string{defaultPath}); err == nil {
		t.Fatalf("Read should fail on bad TAR archive")
	}
}

// tarToFiles() tests.
func TestTarToFiles(t *testing.T) {
	t.Parallel()

	testConfig := &docker.Config{
		ClientGetter: func(...client.Opt) (docker.Client, error) {
			return &docker.FakeClient{
				CopyFromContainerF: func(_ context.Context, _, _ string) (io.ReadCloser, dockertypes.ContainerPathStat, error) {
					return io.NopCloser(testTar(t)), dockertypes.ContainerPathStat{}, nil
				},
			}, nil
		},
	}

	testClient, err := testConfig.New()
	if err != nil {
		t.Fatalf("Unexpected error creating test client: %v", err)
	}

	filesFromArchive, err := testClient.Read("foo", []string{defaultPath})
	if err != nil {
		t.Fatalf("Unexpected error reading from container: %v", err)
	}

	expectedFiles := []*types.File{
		{
			Path:    "/foo",
			Content: "foo\n",
			Mode:    defaultMode,
			User:    "1000",
			Group:   "1000",
		},
	}

	if diff := cmp.Diff(filesFromArchive, expectedFiles); diff != "" {
		t.Fatalf("Got unexpected files: %s", diff)
	}
}

// filesToTar() tests.
//
//nolint:funlen // Just lengthy test.
func TestFilesToTar(t *testing.T) {
	t.Parallel()

	testUser := "test"

	testFile := &types.File{
		Content: "foo\n",
		Mode:    defaultMode,
		Path:    defaultPath,
		User:    testUser,
		Group:   testUser,
	}

	testConfig := &docker.Config{
		ClientGetter: func(...client.Opt) (docker.Client, error) {
			return &docker.FakeClient{
				CopyToContainerF: func(
					ctx context.Context,
					container,
					path string,
					r io.Reader,
					options dockertypes.CopyToContainerOptions,
				) error {
					tr := tar.NewReader(r)

					header, err := tr.Next()
					if err == io.EOF {
						t.Fatalf("At least one file should be found in TAR archive")
					}

					if header.Name != testFile.Path {
						t.Fatalf("Bad file name, expected %s, got %s", testFile.Path, header.Name)
					}

					if header.Mode != testFile.Mode {
						t.Fatalf("Bad file mode, expected %d, got %d", testFile.Mode, header.Mode)
					}

					if header.ModTime.IsZero() {
						t.Fatalf("Modification time in file should be set to current time")
					}

					if header.Uname != testUser {
						t.Fatalf("Expecter uname to be %s, got %s", testUser, header.Uname)
					}

					if header.Gname != testUser {
						t.Fatalf("Expected gname to be %s, got %s", testUser, header.Gname)
					}

					return nil
				},
			}, nil
		},
	}

	testClient, err := testConfig.New()
	if err != nil {
		t.Fatalf("Unexpected error creating test client: %v", err)
	}

	if err := testClient.Copy("", []*types.File{testFile}); err != nil {
		t.Fatalf("Unexpected error while copying: %v", err)
	}
}

func TestFilesToTarNumericUserGroup(t *testing.T) {
	t.Parallel()

	expectedOwnerID := 1001

	testConfig := &docker.Config{
		ClientGetter: func(...client.Opt) (docker.Client, error) {
			return &docker.FakeClient{
				CopyToContainerF: func(
					ctx context.Context,
					container,
					path string,
					r io.Reader,
					options dockertypes.CopyToContainerOptions,
				) error {
					tr := tar.NewReader(r)

					header, err := tr.Next()
					if err == io.EOF {
						t.Fatalf("At least one file should be found in TAR archive")
					}

					if header.Uid != expectedOwnerID {
						t.Fatalf("Expecter uid to be %d, got %d", expectedOwnerID, header.Uid)
					}

					if header.Gid != expectedOwnerID {
						t.Fatalf("Expected gid to be %d, got %d", expectedOwnerID, header.Gid)
					}

					return nil
				},
			}, nil
		},
	}

	testClient, err := testConfig.New()
	if err != nil {
		t.Fatalf("Unexpected error creating test client: %v", err)
	}

	testFile := &types.File{
		Content: "foo\n",
		Mode:    defaultMode,
		Path:    defaultPath,
		User:    strconv.Itoa(expectedOwnerID),
		Group:   strconv.Itoa(expectedOwnerID),
	}

	if err := testClient.Copy("", []*types.File{testFile}); err != nil {
		t.Fatalf("Unexpected error while copying: %v", err)
	}
}

// Create() tests.
func TestCreatePullImageFail(t *testing.T) {
	t.Parallel()

	testConfig := &docker.Config{
		ClientGetter: func(...client.Opt) (docker.Client, error) {
			return &docker.FakeClient{
				ImageListF: func(ctx context.Context, options dockertypes.ImageListOptions) ([]dockertypes.ImageSummary, error) {
					return []dockertypes.ImageSummary{}, fmt.Errorf("runtime error")
				},
			}, nil
		},
	}

	testClient, err := testConfig.New()
	if err != nil {
		t.Fatalf("Unexpected error creating test client: %v", err)
	}

	if _, err := testClient.Create(&types.ContainerConfig{}); err == nil {
		t.Fatalf("Should fail when runtime error occurs")
	}
}

func TestCreateSetUser(t *testing.T) {
	t.Parallel()

	testContainerConfig := &types.ContainerConfig{
		User: "test",
	}

	testConfig := &docker.Config{
		ClientGetter: func(...client.Opt) (docker.Client, error) {
			return &docker.FakeClient{
				ContainerCreateF: func(
					_ context.Context,
					config *containertypes.Config,
					_ *containertypes.HostConfig,
					_ *networktypes.NetworkingConfig,
					_ *v1.Platform,
					_ string,
				) (containertypes.ContainerCreateCreatedBody, error) {
					if config.User != testContainerConfig.User {
						t.Fatalf("Configured user should be %q, got %q", testContainerConfig.User, config.User)
					}

					return containertypes.ContainerCreateCreatedBody{}, nil
				},
				ImagePullF: func(ctx context.Context, ref string, options dockertypes.ImagePullOptions) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("")), nil
				},
				ImageListF: func(ctx context.Context, options dockertypes.ImageListOptions) ([]dockertypes.ImageSummary, error) {
					return []dockertypes.ImageSummary{}, nil
				},
			}, nil
		},
	}

	testClient, err := testConfig.New()
	if err != nil {
		t.Fatalf("Unexpected error creating test client: %v", err)
	}

	if _, err := testClient.Create(testContainerConfig); err != nil {
		t.Fatalf("Create should succeed, got: %v", err)
	}
}

func TestCreateSetUserGroup(t *testing.T) {
	t.Parallel()

	testContainerConfig := &types.ContainerConfig{
		User:  "test",
		Group: "bar",
	}

	expectedUser := fmt.Sprintf("%s:%s", testContainerConfig.User, testContainerConfig.Group)

	testConfig := &docker.Config{
		ClientGetter: func(...client.Opt) (docker.Client, error) {
			return &docker.FakeClient{
				ContainerCreateF: func(
					_ context.Context,
					config *containertypes.Config,
					_ *containertypes.HostConfig,
					_ *networktypes.NetworkingConfig,
					_ *v1.Platform,
					_ string,
				) (containertypes.ContainerCreateCreatedBody, error) {
					if config.User != expectedUser {
						t.Fatalf("Configured user should be %q, got %q", expectedUser, config.User)
					}

					return containertypes.ContainerCreateCreatedBody{}, nil
				},
				ImagePullF: func(ctx context.Context, ref string, options dockertypes.ImagePullOptions) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("")), nil
				},
				ImageListF: func(ctx context.Context, options dockertypes.ImageListOptions) ([]dockertypes.ImageSummary, error) {
					return []dockertypes.ImageSummary{}, nil
				},
			}, nil
		},
	}

	testClient, err := testConfig.New()
	if err != nil {
		t.Fatalf("Unexpected error creating test client: %v", err)
	}

	if _, err := testClient.Create(testContainerConfig); err != nil {
		t.Fatalf("Create should succeed, got: %v", err)
	}
}

func TestCreateRuntimeFail(t *testing.T) {
	t.Parallel()

	testConfig := &docker.Config{
		ClientGetter: func(...client.Opt) (docker.Client, error) {
			return &docker.FakeClient{
				ContainerCreateF: func(
					_ context.Context,
					_ *containertypes.Config,
					_ *containertypes.HostConfig,
					_ *networktypes.NetworkingConfig,
					_ *v1.Platform,
					_ string,
				) (containertypes.ContainerCreateCreatedBody, error) {
					return containertypes.ContainerCreateCreatedBody{}, fmt.Errorf("runtime error")
				},
				ImagePullF: func(ctx context.Context, ref string, options dockertypes.ImagePullOptions) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("")), nil
				},
				ImageListF: func(ctx context.Context, options dockertypes.ImageListOptions) ([]dockertypes.ImageSummary, error) {
					return []dockertypes.ImageSummary{}, nil
				},
			}, nil
		},
	}

	testClient, err := testConfig.New()
	if err != nil {
		t.Fatalf("Unexpected error creating test client: %v", err)
	}

	if _, err := testClient.Create(&types.ContainerConfig{}); err == nil {
		t.Fatalf("Should fail when runtime error occurs")
	}
}

// DefaultConfig() tests.
func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	if docker.DefaultConfig().Host != client.DefaultDockerHost {
		t.Fatalf("Host should be set to %s, got %s", client.DefaultDockerHost, docker.DefaultConfig().Host)
	}
}

// GetAddress() tests.
func TestGetAddressNilConfig(t *testing.T) {
	t.Parallel()

	var c *docker.Config

	if a := c.GetAddress(); a != client.DefaultDockerHost {
		t.Fatalf("Expected %q, got %q", client.DefaultDockerHost, a)
	}
}

func TestGetAddressEmptyConfig(t *testing.T) {
	t.Parallel()

	c := &docker.Config{}

	if a := c.GetAddress(); a != client.DefaultDockerHost {
		t.Fatalf("Expected %q, got %q", client.DefaultDockerHost, a)
	}
}

func TestGetAddress(t *testing.T) {
	t.Parallel()

	expectedAddress := "foo"
	c := &docker.Config{
		Host: expectedAddress,
	}

	if a := c.GetAddress(); a != expectedAddress {
		t.Fatalf("Expected %q, got %q", expectedAddress, a)
	}
}

// convertContainerConfig() tests.
func TestConvertContainerConfigEnvVariables(t *testing.T) {
	t.Parallel()

	testContainerConfig := &types.ContainerConfig{
		Env: map[string]string{"foo": "bar"},
	}

	expectedEnvVariables := []string{"foo=bar"}

	testConfig := &docker.Config{
		ClientGetter: func(...client.Opt) (docker.Client, error) {
			return &docker.FakeClient{
				ContainerCreateF: func(
					ctx context.Context,
					config *containertypes.Config,
					hostConfig *containertypes.HostConfig,
					networkingConfig *networktypes.NetworkingConfig,
					platform *v1.Platform,
					containerName string,
				) (containertypes.ContainerCreateCreatedBody, error) {
					if !reflect.DeepEqual(config.Env, expectedEnvVariables) {
						t.Fatalf("Configured environment variables should be included in container configuration")
					}

					return containertypes.ContainerCreateCreatedBody{}, nil
				},
			}, nil
		},
	}

	testClient, err := testConfig.New()
	if err != nil {
		t.Fatalf("Unexpected error creating test client: %v", err)
	}

	if _, err := testClient.Create(testContainerConfig); err != nil {
		t.Fatalf("Unexpected error creating test container: %v", err)
	}
}
