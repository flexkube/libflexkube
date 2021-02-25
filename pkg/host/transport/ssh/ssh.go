// Package ssh is a transport.Interface implementation, which forwards
// given addresses over specified SSH host.
package ssh

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/pkg/host/transport"
)

const (
	// SSHAuthSockEnv is environment variable name used for connecting to ssh-agent.
	SSHAuthSockEnv = "SSH_AUTH_SOCK"
)

// Config represents SSH transport configuration.
//
// All fields are required. Use BuildConfig to pre-fill the configuration with default values.
type Config struct {
	// Address is a hostname or IP address which should be used for connection.
	Address string `json:"address,omitempty"`

	// Port defines which port should be used for SSH connection.
	Port int `json:"port,omitempty"`

	// User defines as which user the connection should authenticate.
	User string `json:"user,omitempty"`

	// Password adds password as one of available authentication methods.
	Password string `json:"password,omitempty"`

	// ConnectionTimeout defines time, after which SSH client gives up single attempt for connecting.
	ConnectionTimeout string `json:"connectionTimeout,omitempty"`

	// RetryTimeout defines after what time connecting should give up, if trying to connect to unreachable
	// host.
	RetryTimeout string `json:"retryTimeout,omitempty"`

	// RetryInterval defines how long to wait between connection attempts.
	RetryInterval string `json:"retryInterval,omitempty"`

	// PrivateKey adds private key as authentication method.
	// It must be defined as valid SSH private key in PEM format.
	PrivateKey string `json:"privateKey,omitempty"`
}

// ssh is an implementation of Transport interface over SSH protocol.
type ssh struct {
	address           string
	user              string
	connectionTimeout time.Duration
	retryTimeout      time.Duration
	retryInterval     time.Duration
	auth              []gossh.AuthMethod
	sshClientGetter   func(network, address string, config *gossh.ClientConfig) (*gossh.Client, error)
}

type sshConnected struct {
	client   dialer
	address  string
	uuid     func() (uuid.UUID, error)
	listener func(string, string) (net.Listener, error)
}

type dialer interface {
	Dial(network, address string) (net.Conn, error)
}

// New validates SSH configuration and returns new instance of transport interface.
func (d *Config) New() (transport.Interface, error) {
	if err := d.Validate(); err != nil {
		return nil, fmt.Errorf("ssh host validation failed: %w", err)
	}

	// Validate checks parsing, so we can skip error checking here.
	ct, _ := time.ParseDuration(d.ConnectionTimeout)
	rt, _ := time.ParseDuration(d.RetryTimeout)
	ri, _ := time.ParseDuration(d.RetryInterval)

	s := &ssh{
		address:           fmt.Sprintf("%s:%d", d.Address, d.Port),
		user:              d.User,
		connectionTimeout: ct,
		retryTimeout:      rt,
		retryInterval:     ri,
		auth:              []gossh.AuthMethod{},
		sshClientGetter:   gossh.Dial,
	}

	if d.Password != "" {
		s.auth = append(s.auth, gossh.Password(d.Password))
	}

	if signer, _ := gossh.ParsePrivateKey([]byte(d.PrivateKey)); d.PrivateKey != "" {
		s.auth = append(s.auth, gossh.PublicKeys(signer))
	}

	// Multiple auth methods might be used, so if SSH_AUTH_SOCK is defined, try to use it
	// automatically. That gives nice user experience, when user don't have to specify any
	// authentication information explicitly.
	if authSock := os.Getenv(SSHAuthSockEnv); authSock != "" {
		authConn, err := net.Dial("unix", authSock)
		if err != nil {
			return nil, fmt.Errorf("dialing SSH agent failed: %w", err)
		}
		// TODO: We should close the authSock with Close() after we finish using it,
		// but it is not trivial at the moment, so we just let the dying process to
		// close it.
		//
		// defer authConn.Close()

		signers, err := agent.NewClient(authConn).Signers()
		if err != nil {
			return nil, fmt.Errorf("getting public keys from SSH agent failed: %w", err)
		}

		s.auth = append(s.auth, gossh.PublicKeys(signers...))
	}

	return s, nil
}

// Validate validates given configuration.
func (d *Config) Validate() error {
	var errors util.ValidateError

	if d.Address == "" {
		errors = append(errors, fmt.Errorf("address must be set"))
	}

	if d.User == "" {
		errors = append(errors, fmt.Errorf("user must be set"))
	}

	if d.Password == "" && d.PrivateKey == "" && os.Getenv(SSHAuthSockEnv) == "" {
		errors = append(errors, fmt.Errorf("at least one authentication method must be available"))
	}

	if d.Port == 0 {
		errors = append(errors, fmt.Errorf("port must be set"))
	}

	errors = append(errors, d.validateDurations()...)

	return errors.Return()
}

func (d *Config) validateDurations() util.ValidateError {
	var errors util.ValidateError

	// Make sure durations are parse-able.
	if _, err := time.ParseDuration(d.ConnectionTimeout); err != nil {
		errors = append(errors, fmt.Errorf("unable to parse connection timeout: %w", err))
	}

	if _, err := time.ParseDuration(d.RetryTimeout); err != nil {
		errors = append(errors, fmt.Errorf("unable to parse retry timeout: %w", err))
	}

	if _, err := time.ParseDuration(d.RetryInterval); err != nil {
		errors = append(errors, fmt.Errorf("unable to parse retry interval: %w", err))
	}

	if _, err := gossh.ParsePrivateKey([]byte(d.PrivateKey)); d.PrivateKey != "" && err != nil {
		errors = append(errors, fmt.Errorf("unable to parse private key: %w", err))
	}

	return errors
}

// Connect opens SSH connection to configured host.
func (d *ssh) Connect() (transport.Connected, error) {
	sshConfig := &gossh.ClientConfig{
		Auth:    d.auth,
		Timeout: d.connectionTimeout,
		User:    d.user,
		// TODO: Add possibility to specify host keys, which should be accepted.
		// Since user may not know the public keys of their server, for convenience,
		// allow insecure host keys.
		//
		// #nosec G106
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	}

	var connection *gossh.Client

	var err error

	start := time.Now()

	// Try until we timeout.
	for time.Since(start) < d.retryTimeout {
		if connection, err = d.sshClientGetter("tcp", d.address, sshConfig); err == nil {
			return newConnected(d.address, connection), nil
		}

		time.Sleep(d.retryInterval)
	}

	return nil, err
}

func newConnected(address string, connection dialer) transport.Connected {
	return &sshConnected{
		client:   connection,
		address:  address,
		uuid:     uuid.NewRandom,
		listener: net.Listen,
	}
}

// ForwardUnixSocket takes remote UNIX socket path as an argument and forwards
// it to the local socket.
func (d *sshConnected) ForwardUnixSocket(path string) (string, error) {
	unixAddr, err := d.randomUnixSocket()
	if err != nil {
		return "", fmt.Errorf("failed generating random socket to listen: %w", err)
	}

	localSock, err := d.listener("unix", unixAddr.String())
	if err != nil {
		return "", fmt.Errorf("unable to listen on address '%s':%w", unixAddr, err)
	}

	path, err = extractPath(path)
	if err != nil {
		return "", fmt.Errorf("failed parsing path %s: %w", path, err)
	}

	// Schedule accepting connections and return.
	go forwardConnection(localSock, d.client, path, "unix")

	return fmt.Sprintf("unix://%s", unixAddr.String()), nil
}

// handleClient is responsible for copying incoming and outgoing data going
// through the forwarded connection.
func handleClient(client io.ReadWriteCloser, remote io.ReadWriteCloser) {
	defer func() {
		if err := client.Close(); err != nil {
			fmt.Printf("failed closing client connection: %v\n", err)
		}

		if err := remote.Close(); err != nil {
			fmt.Printf("closing remote: %v\n", err)
		}
	}()

	chDone := make(chan bool)

	// Start remote -> local data transfer.
	go func() {
		if _, err := io.Copy(client, remote); err != nil {
			fmt.Printf("error while copy remote->local: %s\n", err)
		}
		chDone <- true
	}()

	// Start local -> remote data transfer.
	go func() {
		if _, err := io.Copy(remote, client); err != nil {
			fmt.Printf("error while copy local->remote: %s\n", err)
		}
		chDone <- true
	}()

	<-chDone
}

// forwardConnection accepts local connections, and forwards them to remote address.
//
// TODO: Should we do some error handling here?
func forwardConnection(l net.Listener, connection dialer, remoteAddress, connectionType string) {
	defer func() {
		if err := l.Close(); err != nil {
			fmt.Printf("failed closing listener: %v\n", err)
		}
	}()

	for {
		// Accept connection from the client.
		c, err := l.Accept()
		if err != nil {
			fmt.Printf("failed to accept connection: %v\n", err)
			// Handle error (and then for example indicate acceptor is down).
			return
		}

		// Open remote connection.
		remoteSock, err := connection.Dial(connectionType, remoteAddress)
		if err != nil {
			fmt.Printf("failed to open remote connection: %v\n", err)

			return
		}

		// Schedule data transfers.
		go handleClient(c, remoteSock)
	}
}

// extractPath parses and verifies, that given URL is unix socket URL
// and returns it's path without the scheme.
func extractPath(path string) (string, error) {
	url, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("unable to parse path %s: %w", path, err)
	}

	if url.Scheme != "unix" {
		return "", fmt.Errorf("forwarding non-unix socket paths is not supported")
	}

	return url.Path, nil
}

// randomUnixSocket generates random abstract UNIX socket, including unique UUID,
// to avoid collisions.
func (d *sshConnected) randomUnixSocket() (*net.UnixAddr, error) {
	// TODO: Rather than connecting again every time ForwardUnixSocket is called
	// we should cache and reuse the connections.
	id, err := d.uuid()
	if err != nil {
		return nil, fmt.Errorf("unable to generate random UUID for abstract UNIX socket: %w", err)
	}

	return &net.UnixAddr{
		Name: fmt.Sprintf("@%s-%s", d.address, id),
		Net:  "unix",
	}, nil
}

// ForwardTCP takes remote TCP address, starts listening on local port and forwards all incoming
// connections to local address to remote address using estabilshed SSH tunnel.
func (d *sshConnected) ForwardTCP(address string) (string, error) {
	if _, _, err := net.SplitHostPort(address); err != nil {
		return "", fmt.Errorf("failed to validate address '%s': %w", address, err)
	}

	localConn, err := d.listener("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("unable to listen on random TCP port: %w", err)
	}

	// Schedule accepting connections and return.
	go forwardConnection(localConn, d.client, address, "tcp")

	return localConn.Addr().String(), nil
}
