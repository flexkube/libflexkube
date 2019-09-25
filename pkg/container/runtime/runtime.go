package runtime

import (
	"io"
	"os"

	"github.com/invidian/etcd-ariadnes-thread/pkg/container/types"
)

// Runtime interface describes universal way of managing containers
// across different container runtimes.
type Runtime interface {
	// Create creates container and returns it's unique identifier
	Create(config *types.ContainerConfig) (string, error)
	// Delete removes the container
	Delete(ID string) error
	// Start starts created container
	Start(ID string) error
	// Status returns status of the container
	Status(ID string) (*types.ContainerStatus, error)
	// Stop takes unique identifier as a parameter and stops the container
	Stop(ID string) error
	// Copy allows to copy TAR archive into the container
	//
	// Docker currently does not allow to copy muliple files over https://github.com/moby/moby/issues/7710
	// It seems kubelet does https://github.com/kubernetes/kubernetes/pull/72641/files
	// TODO consider batching copying and reading
	Copy(ID string, dstPath string, content io.Reader) error
	// Read allows to read file in TAR archive format from container
	// TODO check if we should return some information about read file
	Read(ID string, srcPath string) (io.ReadCloser, error)

	// Stat returns os.FileMode from inside the container
	// TODO this should be improved
	Stat(ID string, path string) (*os.FileMode, error)
}

// RuntimeConfig defines interface for runtime configuration. Since some feature are generic to runtime,
// this interface make sure that other parts of the system are compatible with it.
type RuntimeConfig interface {
	GetAddress() string
	SetAddress(string)
	New() (Runtime, error)
}
