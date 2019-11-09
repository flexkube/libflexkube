package host

import (
	"fmt"

	"github.com/flexkube/libflexkube/pkg/host/transport"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
)

// Host allows to forward TCP ports, UNIX sockets to local machine to establish
// communication with remote daemons.
type Host struct {
	DirectConfig *direct.DirectConfig `json:"direct,omitempty" yaml:"direct,omitempty"`
	SSHConfig    *ssh.SSHConfig       `json:"ssh,omitempty" yaml:"ssh,omitempty"`
}

type host struct {
	transportConfig transport.TransportConfig
}

type hostConnected struct {
	transport transport.Transport
}

func New(h *Host) (*host, error) {
	if err := h.Validate(); err != nil {
		return nil, fmt.Errorf("host configuration validation failed: %w", err)
	}
	// TODO that seems ugly, is there a better way to generalize it?
	var t transport.TransportConfig
	if h.DirectConfig != nil {
		t = h.DirectConfig
	}
	if h.SSHConfig != nil {
		t = h.SSHConfig
	}
	return &host{
		transportConfig: t,
	}, nil
}

// Validate validates host configuration
func (h *Host) Validate() error {
	if err := h.DirectConfig.Validate(); err != nil {
		return fmt.Errorf("direct config validation failed: %w", err)
	}
	if h.DirectConfig != nil && h.SSHConfig != nil {
		return fmt.Errorf("host must have only one transport method defined")
	}
	if h.DirectConfig == nil && h.SSHConfig == nil {
		return fmt.Errorf("host must have transport method defined")
	}
	return nil
}

// selectTransport returns transport protocol configured for container
//
// It returns error if transport protocol configuration is invalid
func (h *host) Connect() (*hostConnected, error) {
	d, err := h.transportConfig.New()
	if err != nil {
		return nil, fmt.Errorf("selecting transport protocol failed: %w", err)
	}

	return &hostConnected{
		transport: d,
	}, nil
}

func (h *hostConnected) ForwardUnixSocket(path string) (string, error) {
	return h.transport.ForwardUnixSocket(path)
}
