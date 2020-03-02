package docker

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"

	dockertypes "github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	networktypes "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/container/types"
)

// New()
func TestNewClient(t *testing.T) {
	// TODO does this kind of simple tests make sense? Integration tests calls the same thing
	// anyway. Or maybe we should simply skip error checking in itegration tests to simplify them?
	c := &Config{}
	if _, err := c.New(); err != nil {
		t.Errorf("Creating new docker client should work, got: %s", err)
	}
}

// getDockerClient()
func TestNewClientWithHost(t *testing.T) {
	config := &Config{
		Host: "unix:///foo.sock",
	}

	c, err := config.getDockerClient()
	if err != nil {
		t.Fatalf("Creating new docker client should work, got: %s", err)
	}

	if dh := c.DaemonHost(); dh != config.Host {
		t.Fatalf("Client with host set should have '%s' as host, got: '%s'", config.Host, dh)
	}
}

// sanitizeImageName()
func TestSanitizeImageName(t *testing.T) {
	e := "foo:latest"

	if g := sanitizeImageName("foo"); g != e {
		t.Fatalf("Expected '%s', got '%s'", e, g)
	}
}

func TestSanitizeImageNameWithTag(t *testing.T) {
	e := "foo:v0.1.0"

	if g := sanitizeImageName(e); g != e {
		t.Fatalf("Expected '%s', got '%s'", e, g)
	}
}

// Status()
func TestStatus(t *testing.T) {
	es := "running"

	d := &docker{
		ctx: context.Background(),
		cli: &FakeClient{
			ContainerInspectF: func(ctx context.Context, id string) (dockertypes.ContainerJSON, error) {
				return dockertypes.ContainerJSON{
					ContainerJSONBase: &dockertypes.ContainerJSONBase{
						State: &dockertypes.ContainerState{
							Status: es,
						},
					},
				}, nil
			},
		},
	}

	s, err := d.Status("foo")
	if err != nil {
		t.Fatalf("Checking for status should succeed, got: %v", err)
	}

	if s.ID == "" {
		t.Fatalf("ID in status of existing container should not be empty")
	}

	if s.Status != es {
		t.Fatalf("Received status should be %s, got %s", es, s.Status)
	}
}

func TestStatusNotFound(t *testing.T) {
	d := &docker{
		ctx: context.Background(),
		cli: &FakeClient{
			ContainerInspectF: func(ctx context.Context, id string) (dockertypes.ContainerJSON, error) {
				return dockertypes.ContainerJSON{}, errdefs.NotFound(fmt.Errorf("not found"))
			},
		},
	}

	s, err := d.Status("foo")
	if err != nil {
		t.Fatalf("Checking for status should succeed, got: %v", err)
	}

	if s.ID != "" {
		t.Fatalf("ID in status of non-existing container should be empty")
	}
}

func TestStatusRuntimeError(t *testing.T) {
	d := &docker{
		ctx: context.Background(),
		cli: &FakeClient{
			ContainerInspectF: func(ctx context.Context, id string) (dockertypes.ContainerJSON, error) {
				return dockertypes.ContainerJSON{}, fmt.Errorf("can't check status of container")
			},
		},
	}

	if _, err := d.Status("foo"); err == nil {
		t.Fatalf("Checking for status should fail")
	}
}

// Copy()
func TestCopyRuntimeError(t *testing.T) {
	d := &docker{
		ctx: context.Background(),
		cli: &FakeClient{
			CopyToContainerF: func(ctx context.Context, id string, path string, content io.Reader, options dockertypes.CopyToContainerOptions) error {
				return fmt.Errorf("Copying failed")
			},
		},
	}

	if err := d.Copy("foo", []*types.File{}); err == nil {
		t.Fatalf("should fail when runtime returns error")
	}
}

// Read()
func TestReadRuntimeError(t *testing.T) {
	p := defaultPath

	d := &docker{
		ctx: context.Background(),
		cli: &FakeClient{
			CopyFromContainerF: func(ctx context.Context, id string, path string) (io.ReadCloser, dockertypes.ContainerPathStat, error) {
				if path != p {
					t.Fatalf("Should read path %s, got %s", p, path)
				}

				return nil, dockertypes.ContainerPathStat{}, fmt.Errorf("Copying failed")
			},
		},
	}

	if _, err := d.Read("foo", []string{p}); err == nil {
		t.Fatalf("should fail when runtime returns error")
	}
}

const (
	defaultMode = 420
	defaultPath = "/foo"
)

func TestRead(t *testing.T) {
	p := defaultPath

	d := &docker{
		ctx: context.Background(),
		cli: &FakeClient{
			CopyFromContainerF: func(ctx context.Context, id string, path string) (io.ReadCloser, dockertypes.ContainerPathStat, error) {
				return ioutil.NopCloser(testTar(t)), dockertypes.ContainerPathStat{
					Name: p,
				}, nil
			},
		},
	}

	fs, err := d.Read("foo", []string{p})
	if err != nil {
		t.Fatalf("Reading should succeed, got: %v", err)
	}

	files := []*types.File{
		{
			Path:    p,
			Content: "foo\n",
			Mode:    defaultMode,
			User:    "1000",
			Group:   "1000",
		},
	}

	if diff := cmp.Diff(files, fs); diff != "" {
		t.Fatalf("Got unexpected files: %s", diff)
	}
}

func TestReadFileMissing(t *testing.T) {
	p := defaultPath

	d := &docker{
		ctx: context.Background(),
		cli: &FakeClient{
			CopyFromContainerF: func(ctx context.Context, id string, path string) (io.ReadCloser, dockertypes.ContainerPathStat, error) {
				return nil, dockertypes.ContainerPathStat{}, nil
			},
		},
	}

	fs, err := d.Read("foo", []string{p})
	if err != nil {
		t.Fatalf("read should succeed, got: %v", err)
	}

	if len(fs) != 0 {
		t.Fatalf("read should not return any files if the file does not exist")
	}
}

func testTar(t *testing.T) io.Reader {
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
	p := defaultPath

	d := &docker{
		ctx: context.Background(),
		cli: &FakeClient{
			CopyFromContainerF: func(ctx context.Context, id string, path string) (io.ReadCloser, dockertypes.ContainerPathStat, error) {
				return ioutil.NopCloser(strings.NewReader("asdasd")), dockertypes.ContainerPathStat{}, nil
			},
		},
	}

	if _, err := d.Read("foo", []string{p}); err == nil {
		t.Fatalf("read should fail on bad TAR archive")
	}
}

// tarToFiles()
func TestTarToFiles(t *testing.T) {
	fs, err := tarToFiles(testTar(t))
	if err != nil {
		t.Fatalf("Reading should succeed, got: %v", err)
	}

	files := []*types.File{
		{
			Content: "foo\n",
			Mode:    defaultMode,
			User:    "1000",
			Group:   "1000",
		},
	}

	if diff := cmp.Diff(files, fs); diff != "" {
		t.Fatalf("Got unexpected files: %s", diff)
	}
}

// filesToTar()
func TestFilesToTar(t *testing.T) {
	tn := "test"
	f := &types.File{
		Content: "foo\n",
		Mode:    defaultMode,
		Path:    defaultPath,
		User:    tn,
		Group:   tn,
	}

	r, err := filesToTar([]*types.File{f})
	if err != nil {
		t.Fatalf("Packing files should succeed, got: %v", err)
	}

	tr := tar.NewReader(r)

	h, err := tr.Next()
	if err == io.EOF {
		t.Fatalf("At least one file should be found in TAR archive")
	}

	if h.Name != f.Path {
		t.Fatalf("Bad file name, expected %s, got %s", f.Path, h.Name)
	}

	if h.Mode != f.Mode {
		t.Fatalf("Bad file mode, expected %d, got %d", f.Mode, h.Mode)
	}

	if h.ModTime.IsZero() {
		t.Fatalf("Modification time in file should be set to current time")
	}

	if h.Uname != tn {
		t.Fatalf("Expecter uname to be %s, got %s", tn, h.Uname)
	}

	if h.Gname != tn {
		t.Fatalf("Expected gname to be %s, got %s", tn, h.Gname)
	}
}

func TestFilesToTarNumericUserGroup(t *testing.T) {
	tn := 1001
	f := &types.File{
		Content: "foo\n",
		Mode:    defaultMode,
		Path:    defaultPath,
		User:    strconv.Itoa(tn),
		Group:   strconv.Itoa(tn),
	}

	r, err := filesToTar([]*types.File{f})
	if err != nil {
		t.Fatalf("Packing files should succeed, got: %v", err)
	}

	tr := tar.NewReader(r)

	h, err := tr.Next()
	if err == io.EOF {
		t.Fatalf("At least one file should be found in TAR archive")
	}

	if h.Uid != tn {
		t.Fatalf("Expecter uid to be %d, got %d", tn, h.Uid)
	}

	if h.Gid != tn {
		t.Fatalf("Expected gid to be %d, got %d", tn, h.Gid)
	}
}

// Create()
func TestCreatePullImageFail(t *testing.T) {
	d := &docker{
		ctx: context.Background(),
		cli: &FakeClient{
			ImageListF: func(ctx context.Context, options dockertypes.ImageListOptions) ([]dockertypes.ImageSummary, error) {
				return []dockertypes.ImageSummary{}, fmt.Errorf("runtime error")
			},
		},
	}

	if _, err := d.Create(&types.ContainerConfig{}); err == nil {
		t.Fatalf("Should fail when runtime error occurs")
	}
}

func TestCreateBuildPortsFail(t *testing.T) {}

func TestCreateSetUser(t *testing.T) {
	c := &types.ContainerConfig{
		User: "test",
	}

	d := &docker{
		ctx: context.Background(),
		cli: &FakeClient{
			ContainerCreateF: func(ctx context.Context, config *containertypes.Config, hostConfig *containertypes.HostConfig, networkingConfig *networktypes.NetworkingConfig, containerName string) (containertypes.ContainerCreateCreatedBody, error) {
				if config.User != c.User {
					t.Fatalf("configured user should be %s, got %s", c.User, config.User)
				}

				return containertypes.ContainerCreateCreatedBody{}, nil
			},
			ImagePullF: func(ctx context.Context, ref string, options dockertypes.ImagePullOptions) (io.ReadCloser, error) {
				return ioutil.NopCloser(strings.NewReader("")), nil
			},
			ImageListF: func(ctx context.Context, options dockertypes.ImageListOptions) ([]dockertypes.ImageSummary, error) {
				return []dockertypes.ImageSummary{}, nil
			},
		},
	}

	if _, err := d.Create(c); err != nil {
		t.Fatalf("Create should succeed, got: %v", err)
	}
}

func TestCreateSetUserGroup(t *testing.T) {
	c := &types.ContainerConfig{
		User:  "test",
		Group: "bar",
	}

	e := fmt.Sprintf("%s:%s", c.User, c.Group)

	d := &docker{
		ctx: context.Background(),
		cli: &FakeClient{
			ContainerCreateF: func(ctx context.Context, config *containertypes.Config, hostConfig *containertypes.HostConfig, networkingConfig *networktypes.NetworkingConfig, containerName string) (containertypes.ContainerCreateCreatedBody, error) {
				if config.User != e {
					t.Fatalf("configured user should be %s, got %s", e, config.User)
				}

				return containertypes.ContainerCreateCreatedBody{}, nil
			},
			ImagePullF: func(ctx context.Context, ref string, options dockertypes.ImagePullOptions) (io.ReadCloser, error) {
				return ioutil.NopCloser(strings.NewReader("")), nil
			},
			ImageListF: func(ctx context.Context, options dockertypes.ImageListOptions) ([]dockertypes.ImageSummary, error) {
				return []dockertypes.ImageSummary{}, nil
			},
		},
	}

	if _, err := d.Create(c); err != nil {
		t.Fatalf("Create should succeed, got: %v", err)
	}
}

func TestCreateRuntimeFail(t *testing.T) {
	d := &docker{
		ctx: context.Background(),
		cli: &FakeClient{
			ContainerCreateF: func(ctx context.Context, config *containertypes.Config, hostConfig *containertypes.HostConfig, networkingConfig *networktypes.NetworkingConfig, containerName string) (containertypes.ContainerCreateCreatedBody, error) {
				return containertypes.ContainerCreateCreatedBody{}, fmt.Errorf("runtime error")
			},
			ImagePullF: func(ctx context.Context, ref string, options dockertypes.ImagePullOptions) (io.ReadCloser, error) {
				return ioutil.NopCloser(strings.NewReader("")), nil
			},
			ImageListF: func(ctx context.Context, options dockertypes.ImageListOptions) ([]dockertypes.ImageSummary, error) {
				return []dockertypes.ImageSummary{}, nil
			},
		},
	}

	if _, err := d.Create(&types.ContainerConfig{}); err == nil {
		t.Fatalf("Should fail when runtime error occurs")
	}
}

func TestCreate(t *testing.T) {}

// DefaultConfig()
func TestDefaultConfig(t *testing.T) {
	if DefaultConfig().Host != client.DefaultDockerHost {
		t.Fatalf("Host should be set to %s, got %s", client.DefaultDockerHost, DefaultConfig().Host)
	}
}
