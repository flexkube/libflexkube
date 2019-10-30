package direct

import (
	"fmt"

	"github.com/invidian/libflexkube/pkg/host/transport"
)

type DirectConfig struct{}

type direct struct{}

// New may in the future validate direct configuration.
func (d *DirectConfig) New() (transport.Transport, error) {
	if err := d.Validate(); err != nil {
		return nil, fmt.Errorf("direct host validation failed: %w", err)
	}
	return &direct{}, nil
}

func (d *DirectConfig) Validate() error {
	return nil
}

func (d *direct) ForwardUnixSocket(path string) (string, error) {
	return path, nil
}
