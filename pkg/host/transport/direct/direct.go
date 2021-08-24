// Package direct is a transport.Interface implementation, which simply
// forwards given addresses "as is", without any modifications.
package direct

import (
	"fmt"
	"net"

	"github.com/flexkube/libflexkube/pkg/host/transport"
)

// Config represents host configuration for direct communication.
//
// Using this struct will use local network and local filesystem.
type Config struct {
	// Dummy field is only user for testing.
	Dummy string `json:"-"`
}

// direct is a initialized struct, which satisfies Transport interface.
type direct struct{}

// New may in the future validate direct configuration.
func (c *Config) New() (transport.Interface, error) {
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

// Connect implements Transport interface.
func (d *direct) Connect() (transport.Connected, error) {
	return d, nil
}

func (d *direct) ForwardTCP(address string) (string, error) {
	if _, _, err := net.SplitHostPort(address); err != nil {
		return "", fmt.Errorf("validating address %q: %w", address, err)
	}

	return address, nil
}
