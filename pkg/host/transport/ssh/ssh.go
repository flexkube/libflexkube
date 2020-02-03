package ssh

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"time"

	"github.com/google/uuid"
	gossh "golang.org/x/crypto/ssh"

	"github.com/flexkube/libflexkube/pkg/host/transport"
)

// Config represents SSH transport configuration
type Config struct {
	Address           string `json:"address"`
	Port              int    `json:"port"`
	User              string `json:"user"`
	Password          string `json:"password,omitempty"`
	ConnectionTimeout string `json:"connectionTimeout"`
	RetryTimeout      string `json:"retryTimeout"`
	RetryInterval     string `json:"retryInterval"`
	PrivateKey        string `json:"privateKey,omitempty"`
}

// ssh is an implementation of Transport interface over SSH protocol
type ssh struct {
	address           string
	user              string
	connectionTimeout time.Duration
	retryTimeout      time.Duration
	retryInterval     time.Duration
	auth              []gossh.AuthMethod
	sshClientGetter   func(network string, address string, config *gossh.ClientConfig) (*gossh.Client, error)
}

type sshConnected struct {
	client   dialer
	address  string
	uuid     func() (uuid.UUID, error)
	listener func(string, string) (net.Listener, error)
}

type dialer interface {
	Dial(network string, address string) (net.Conn, error)
}

// New creates new instance of ssh struct
func (d *Config) New() (transport.Interface, error) {
	if err := d.Validate(); err != nil {
		return nil, fmt.Errorf("ssh host validation failed: %w", err)
	}

	// Validate checks parsing, so we can skip error checking here
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

	if d.PrivateKey != "" {
		signer, _ := gossh.ParsePrivateKey([]byte(d.PrivateKey))
		s.auth = append(s.auth, gossh.PublicKeys(signer))
	}

	return s, nil
}

// Validate validates given configuration and returns on first encountered error
func (d *Config) Validate() error {
	if d.Address == "" {
		return fmt.Errorf("address must be set")
	}

	if d.User == "" {
		return fmt.Errorf("user must be set")
	}

	if d.Password == "" && d.PrivateKey == "" {
		return fmt.Errorf("either password or private key must be set for authentication")
	}

	if d.ConnectionTimeout == "" {
		return fmt.Errorf("connection timeout must be set")
	}

	if d.RetryTimeout == "" {
		return fmt.Errorf("retry timeout must be set")
	}

	if d.RetryInterval == "" {
		return fmt.Errorf("retry interval must be set")
	}

	if d.Port == 0 {
		return fmt.Errorf("port must be set")
	}

	// Make sure durations are parse-able.
	if _, err := time.ParseDuration(d.ConnectionTimeout); err != nil {
		return fmt.Errorf("unable to parse connection timeout: %w", err)
	}

	if _, err := time.ParseDuration(d.RetryTimeout); err != nil {
		return fmt.Errorf("unable to parse retry timeout: %w", err)
	}

	if _, err := time.ParseDuration(d.RetryInterval); err != nil {
		return fmt.Errorf("unable to parse retry interval: %w", err)
	}

	if d.PrivateKey != "" {
		if _, err := gossh.ParsePrivateKey([]byte(d.PrivateKey)); err != nil {
			return fmt.Errorf("unable to parse private key: %w", err)
		}
	}

	return nil
}

func (d *ssh) Connect() (transport.Connected, error) {
	sshConfig := &gossh.ClientConfig{
		Auth:    d.auth,
		Timeout: d.connectionTimeout,
		User:    d.user,
		// TODO add possibility to specify host keys, which should be accepted.
		// Since user may not know the public keys of their server, for convenience,
		// allow insecure host keys.
		//
		// #nosec G106
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	}

	var connection *gossh.Client

	var err error

	start := time.Now()

	// Try until we timeout
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
// through the forwarded connection
func handleClient(client net.Conn, remote io.ReadWriter) {
	defer client.Close()

	chDone := make(chan bool)

	// Start remote -> local data transfer
	go func() {
		if _, err := io.Copy(client, remote); err != nil {
			fmt.Printf("error while copy remote->local: %s\n", err)
		}
		chDone <- true
	}()

	// Start local -> remote data transfer
	go func() {
		if _, err := io.Copy(remote, client); err != nil {
			fmt.Printf("error while copy local->remote: %s\n", err)
		}
		chDone <- true
	}()

	<-chDone
}

// forwardConnection accepts local connections, and forwards them to remote address
//
// TODO should we do some error handling here?
func forwardConnection(l net.Listener, connection dialer, remoteAddress string, connectionType string) {
	defer l.Close()

	for {
		// Accept connection from the client.
		c, err := l.Accept()
		if err != nil {
			fmt.Printf("failed to accept connection: %v\n", err)
			// handle error (and then for example indicate acceptor is down)
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
	// TODO rather than connecting again every time ForwardUnixSocket is called
	// we should cache and reuse the connections
	id, err := d.uuid()
	if err != nil {
		return nil, fmt.Errorf("unable to generate random UUID for abstract UNIX socket: %w", err)
	}

	return &net.UnixAddr{
		Name: fmt.Sprintf("@%s-%s", d.address, id),
		Net:  "unix",
	}, nil
}

func (d *sshConnected) ForwardTCP(address string) (string, error) {
	localConn, err := d.listener("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("unable to listen on random TCP port: %w", err)
	}

	// Schedule accepting connections and return.
	go forwardConnection(localConn, d.client, address, "tcp")

	return localConn.Addr().String(), nil
}
