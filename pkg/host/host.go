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
type Host struct {
	DirectConfig *direct.Config `json:"direct,omitempty"`
	SSHConfig    *ssh.Config    `json:"ssh,omitempty"`
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
		return nil, fmt.Errorf("host configuration validation failed: %w", err)
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
	var errors util.ValidateError

	if err := h.DirectConfig.Validate(); err != nil {
		errors = append(errors, fmt.Errorf("direct config validation failed: %w", err))
	}

	if h.DirectConfig != nil && h.SSHConfig != nil {
		errors = append(errors, fmt.Errorf("host must have only one transport method defined"))
	}

	if h.DirectConfig == nil && h.SSHConfig == nil {
		errors = append(errors, fmt.Errorf("host must have transport method defined"))
	}

	if h.SSHConfig != nil {
		if err := h.SSHConfig.Validate(); err != nil {
			errors = append(errors, fmt.Errorf("host ssh config invalid: %w", err))
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
		return nil, fmt.Errorf("connecting failed: %w", err)
	}

	return &hostConnected{
		transport: c,
	}, nil
}

// ForwardUnixSocket forwards given unix socket path using configured transport method.
func (h *hostConnected) ForwardUnixSocket(path string) (string, error) {
	return h.transport.ForwardUnixSocket(path)
}

func (h *hostConnected) ForwardTCP(address string) (string, error) {
	return h.transport.ForwardTCP(address)
}

// BuildConfig merges values from both host objects.
func BuildConfig(h Host, d Host) Host {
	if h.DirectConfig == nil && h.SSHConfig == nil && d.SSHConfig == nil {
		return Host{
			DirectConfig: &direct.Config{},
		}
	}

	h.SSHConfig = ssh.BuildConfig(h.SSHConfig, d.SSHConfig)

	return h
}
