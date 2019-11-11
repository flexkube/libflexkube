package direct

import (
	"fmt"

	"github.com/flexkube/libflexkube/pkg/host/transport"
)

// Config represents host configuration for direct communication.
//
// Using this struct will use local network and local filesystem.
type Config struct{}

// direct is a initialized struct, which satisfies Transport interface.
type direct struct{}

// New may in the future validate direct configuration.
func (c *Config) New() (transport.Transport, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("direct host validation failed: %w", err)
	}
	return &direct{}, nil
}

// Validate validates Config struct.
//
// Currently it's not doing anything, but it's here for compatibility purposes
// with other types. In the future some validation rules may be added.
func (c *Config) Validate() error {
	return nil
}

// ForwardUnixSocket returns forwarded UNIX socket.
//
// Given that direct operates on local filesystem, it simply returns given path.
//
// TODO perhaps try to connect to given socket to see if it exists, we have permissions
// etc to fail early?
func (d *direct) ForwardUnixSocket(path string) (string, error) {
	return path, nil
}
