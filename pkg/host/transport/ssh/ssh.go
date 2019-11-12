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

type Config struct {
	Address           string `json:"address" yaml:"address"`
	Port              int    `json:"port" yaml:"port"`
	User              string `json:"user" yaml:"user"`
	Password          string `json:"password,omitempty" yaml:"password,omitempty"`
	ConnectionTimeout string `json:"connectionTimeout" yaml:"connectionTimeout"`
	PrivateKey        string `json:"privateKey,omitempty" yaml:"privateKey,omitempty"`
}

type ssh struct {
	address           string
	user              string
	connectionTimeout time.Duration
	auth              []gossh.AuthMethod
}

// New may in the future validate ssh configuration.
func (d *Config) New() (transport.Transport, error) {
	if err := d.Validate(); err != nil {
		return nil, fmt.Errorf("ssh host validation failed: %w", err)
	}

	// Validate checks parsing, so we can skip error checking here
	ct, _ := time.ParseDuration(d.ConnectionTimeout)

	s := &ssh{
		address:           fmt.Sprintf("%s:%d", d.Address, d.Port),
		user:              d.User,
		connectionTimeout: ct,
		auth:              []gossh.AuthMethod{},
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
	if d.Port == 0 {
		return fmt.Errorf("port must be set")
	}

	// Make sure duration is parse-able
	if _, err := time.ParseDuration(d.ConnectionTimeout); err != nil {
		return fmt.Errorf("unable to parse connection timeout: %w", err)
	}

	if d.PrivateKey != "" {
		if _, err := gossh.ParsePrivateKey([]byte(d.PrivateKey)); err != nil {
			return fmt.Errorf("unable to parse private key: %w", err)
		}
	}

	return nil
}

func (d *ssh) ForwardUnixSocket(path string) (string, error) {
	sshConfig := &gossh.ClientConfig{
		Auth:    d.auth,
		Timeout: d.connectionTimeout,
		User:    d.user,
		// TODO add possibility to specify host keys, which should be accepted
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	}

	// TODO Make those intervals configurable
	// TODO alternatively, we could move connecting part to separated method
	// and let user take care of it.
	retryTimeout, _ := time.ParseDuration("60s")
	retryInterval, _ := time.ParseDuration("1s")
	var connection *gossh.Client
	var err error
	start := time.Now()

	// Try until we timeout
	for time.Since(start) < retryTimeout {
		connection, err = gossh.Dial("tcp", d.address, sshConfig)
		if err == nil {
			break
		}
		time.Sleep(retryInterval)
	}
	if err != nil {
		return "", fmt.Errorf("failed to open SSH connection: %w", err)
	}

	url, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("unable to parse path %s: %w", path, err)
	}
	if url.Scheme != "unix" {
		return "", fmt.Errorf("forwarding non-unix socket paths is not supported")
	}

	// Generate new UUID for every connection, to make sure we don't get "address already in use" error
	// TODO rather than connecting again every time ForwardUnixSocket is called
	// we should cache and reuse the connections
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("unable to generate random UUID for abstract UNIX socket: %w", err)
	}

	localAddr := fmt.Sprintf("@%s-%s", d.address, id)
	localSock, err := net.ListenUnix("unix", &net.UnixAddr{localAddr, "unix"})
	if err != nil {
		return "", fmt.Errorf("unable to listen on address '%s':%w", localAddr, err)
	}

	// For every listener spawn the following routine
	go func(l net.Listener, remote string) {
		defer localSock.Close()

		for {
			c, err := l.Accept()
			if err != nil {
				fmt.Printf("failed to accept connection: %w\n", err)
				// handle error (and then for example indicate acceptor is down)
				return
			}
			remoteSock, err := connection.Dial("unix", remote)
			if err != nil {
				fmt.Printf("failed to open remote connection: %w\n", err)
				return
			}

			go handleClient(c, remoteSock)
		}
	}(localSock, url.Path)

	return fmt.Sprintf("unix://%s", localAddr), nil
}

func handleClient(client net.Conn, remote net.Conn) {
	defer client.Close()
	chDone := make(chan bool)

	// Start remote -> local data transfer
	go func() {
		_, err := io.Copy(client, remote)
		if err != nil {
			fmt.Printf("error while copy remote->local: %s\n", err)
		}
		chDone <- true
	}()

	// Start local -> remote data transfer
	go func() {
		_, err := io.Copy(remote, client)
		if err != nil {
			fmt.Printf("error while copy local->remote: %s\n", err)
		}
		chDone <- true
	}()

	<-chDone
}
