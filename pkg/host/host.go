// Package host collects all transport interface implementations and provides an
// unified configuration interface for these.
package host

import (
	"fmt"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/pkg/host/transport"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
)

// Host allows to forward TCP ports, UNIX sockets to local machine to establish
// communication with remote daemons.
//
// Exactly one transport method must be configured.
type Host struct {
	// DirectConfig indicates, that no forwarding should occur, so addresses will
	// be returned as given.
	DirectConfig *direct.Config `json:"direct,omitempty"`

	// SSHConfig configures given addresses to be forwarded using SSH tunneling.
	SSHConfig *ssh.Config `json:"ssh,omitempty"`
}

type host struct {
	transport transport.Interface
}

type hostConnected struct {
	transport transport.Connected
}

// New validates Host configuration and sets configured transport method.
func (h *Host) New() (transport.Interface, error) {
	if err := h.Validate(); err != nil {
		return nil, fmt.Errorf("validating host configuration: %w", err)
	}

	// TODO that seems ugly, is there a better way to generalize it?
	var t transport.Interface

	if h.DirectConfig != nil {
		t, _ = h.DirectConfig.New()
	}

	if h.SSHConfig != nil {
		t, _ = h.SSHConfig.New()
	}

	return &host{
		transport: t,
	}, nil
}

// Validate validates host configuration.
func (h *Host) Validate() error {
	var errors util.ValidateErrors

	if err := h.DirectConfig.Validate(); err != nil {
		errors = append(errors, fmt.Errorf("validating direct config: %w", err))
	}

	if h.DirectConfig != nil && h.SSHConfig != nil {
		errors = append(errors, fmt.Errorf("host must have only one transport method defined"))
	}

	if h.DirectConfig == nil && h.SSHConfig == nil {
		errors = append(errors, fmt.Errorf("host must have transport method defined"))
	}

	if h.SSHConfig != nil {
		if err := h.SSHConfig.Validate(); err != nil {
			errors = append(errors, fmt.Errorf("validating SSH config: %w", err))
		}
	}

	return errors.Return()
}

// selectTransport returns transport protocol configured for container.
//
// It returns error if transport protocol configuration is invalid.
func (h *host) Connect() (transport.Connected, error) {
	c, err := h.transport.Connect()
	if err != nil {
		return nil, fmt.Errorf("connecting: %w", err)
	}

	return &hostConnected{
		transport: c,
	}, nil
}

// ForwardUnixSocket forwards given unix socket path using configured transport method and returns
// local unix socket address.
func (h *hostConnected) ForwardUnixSocket(path string) (string, error) {
	return h.transport.ForwardUnixSocket(path)
}

// ForwardTCP forwards given TCP address using configured transport method and returns local
// address with port.
func (h *hostConnected) ForwardTCP(address string) (string, error) {
	return h.transport.ForwardTCP(address)
}

// BuildConfig merges values from both host objects. This is a helper method used for building hierarchical
// configuration.
func BuildConfig(config, defaults Host) Host {
	// If config has no direct config configured or has SSH config configured, build SSH configuration.
	if (config.DirectConfig == nil && defaults.SSHConfig != nil) || config.SSHConfig != nil {
		config.SSHConfig = ssh.BuildConfig(config.SSHConfig, defaults.SSHConfig)
	}

	// If config has nothing configured and default has no SSH configuration configured,
	// return direct config as a default.
	if config.DirectConfig == nil && config.SSHConfig == nil && defaults.SSHConfig == nil {
		return Host{
			DirectConfig: &direct.Config{},
		}
	}

	return config
}
