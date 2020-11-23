package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const (
	expectedMessage  = "foo"
	expectedResponse = "bar"
	authMethods      = 1
)

// This test may access SSHAuthSockEnv environment variable,
// which is global variable, so to keep things stable, don't run it in parallel.
//
//nolint:paralleltest
func TestNew(t *testing.T) {
	if err := os.Unsetenv(SSHAuthSockEnv); err != nil {
		t.Fatalf("failed unsetting environment variable %q: %v", SSHAuthSockEnv, err)
	}

	c := &Config{
		Address:           "localhost",
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "30s",
		RetryTimeout:      "60s",
		RetryInterval:     "1s",
		Port:              Port,
		PrivateKey:        generateRSAPrivateKey(t),
	}
	if _, err := c.New(); err != nil {
		t.Fatalf("creating new SSH object should succeed, got: %s", err)
	}
}

// This test may access SSHAuthSockEnv environment variable,
// which is global variable, so to keep things stable, don't run it in parallel.
//
//nolint:paralleltest
func TestNewSetPassword(t *testing.T) {
	if err := os.Unsetenv(SSHAuthSockEnv); err != nil {
		t.Fatalf("failed unsetting environment variable %q: %v", SSHAuthSockEnv, err)
	}

	c := &Config{
		Address:           "localhost",
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "30s",
		RetryTimeout:      "60s",
		RetryInterval:     "1s",
		Port:              Port,
	}

	s, err := c.New()
	if err != nil {
		t.Fatalf("creating new SSH object should succeed, got: %s", err)
	}

	if len(s.(*ssh).auth) != authMethods {
		t.Fatalf("when Password field is set, object should include one auth method")
	}
}

// This test may access SSHAuthSockEnv environment variable,
// which is global variable, so to keep things stable, don't run it in parallel.
//
//nolint:paralleltest
func TestNewSetPrivateKey(t *testing.T) {
	if err := os.Unsetenv(SSHAuthSockEnv); err != nil {
		t.Fatalf("failed unsetting environment variable %q: %v", SSHAuthSockEnv, err)
	}

	c := &Config{
		Address:           "localhost",
		User:              "root",
		ConnectionTimeout: "30s",
		RetryTimeout:      "60s",
		RetryInterval:     "1s",
		Port:              Port,
		PrivateKey:        generateRSAPrivateKey(t),
	}

	s, err := c.New()
	if err != nil {
		t.Fatalf("creating new SSH object should succeed, got: %s", err)
	}

	if len(s.(*ssh).auth) != authMethods {
		t.Fatalf("when PrivateKey field is set, object should include one auth method")
	}
}

func TestNewValidate(t *testing.T) {
	t.Parallel()

	c := &Config{}
	if _, err := c.New(); err == nil {
		t.Fatalf("creating new SSH object should validate it")
	}
}

// Validate() tests.
func TestValidateRequireAddress(t *testing.T) {
	t.Parallel()

	c := &Config{
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "30s",
		RetryTimeout:      "60s",
		RetryInterval:     "1s",
		Port:              Port,
	}
	if err := c.Validate(); err == nil {
		t.Fatalf("validating SSH configuration should require address field")
	}
}

// This test may access SSHAuthSockEnv environment variable,
// which is global variable, so to keep things stable, don't run it in parallel.
//
//nolint:paralleltest
func TestValidateRequireAuth(t *testing.T) {
	if err := os.Unsetenv(SSHAuthSockEnv); err != nil {
		t.Fatalf("failed unsetting environment variable %q: %v", SSHAuthSockEnv, err)
	}

	c := &Config{
		Address:           "localhost",
		User:              "root",
		ConnectionTimeout: "30s",
		RetryTimeout:      "60s",
		RetryInterval:     "1s",
		Port:              Port,
	}
	if err := c.Validate(); err == nil {
		t.Fatalf("validating SSH configuration should require either password or private key")
	}
}

func TestValidateRequireUser(t *testing.T) {
	t.Parallel()

	c := &Config{
		Address:           "localhost",
		Password:          "foo",
		ConnectionTimeout: "30s",
		RetryTimeout:      "60s",
		RetryInterval:     "1s",
		Port:              Port,
	}
	if err := c.Validate(); err == nil {
		t.Fatalf("validating SSH configuration should require user field")
	}
}

func TestValidateRequireConnectionTimeout(t *testing.T) {
	t.Parallel()

	c := &Config{
		Address:       "localhost",
		User:          "root",
		Password:      "foo",
		RetryTimeout:  "60s",
		RetryInterval: "1s",
		Port:          Port,
	}
	if err := c.Validate(); err == nil {
		t.Fatalf("validating SSH configuration should require connection timeout field")
	}
}

func TestValidateRequireRetryTimeout(t *testing.T) {
	t.Parallel()

	c := &Config{
		Address:           "localhost",
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "30s",
		RetryInterval:     "1s",
		Port:              Port,
	}
	if err := c.Validate(); err == nil {
		t.Fatalf("validating SSH configuration should require retry timeout field")
	}
}

func TestValidateRequireRetryInterval(t *testing.T) {
	t.Parallel()

	c := &Config{
		Address:           "localhost",
		User:              "root",
		Password:          "foo",
		Port:              Port,
		ConnectionTimeout: "30s",
		RetryTimeout:      "60s",
	}
	if err := c.Validate(); err == nil {
		t.Fatalf("validating SSH configuration should require retry interval field")
	}
}

func TestValidateRequirePort(t *testing.T) {
	t.Parallel()

	c := &Config{
		Address:           "localhost",
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "30s",
		RetryTimeout:      "60s",
		RetryInterval:     "1s",
	}
	if err := c.Validate(); err == nil {
		t.Fatalf("validating SSH configuration should require port field")
	}
}

func TestValidateParseConnectionTimeout(t *testing.T) {
	t.Parallel()

	c := &Config{
		Address:           "localhost",
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "doh",
		RetryTimeout:      "60s",
		RetryInterval:     "1s",
		Port:              Port,
	}
	if err := c.Validate(); err == nil {
		t.Fatalf("validating SSH configuration should parse connection timeout")
	}
}

func TestValidateParseRetryTimeout(t *testing.T) {
	t.Parallel()

	c := &Config{
		Address:           "localhost",
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "30s",
		RetryTimeout:      "doh",
		RetryInterval:     "1s",
		Port:              Port,
	}
	if err := c.Validate(); err == nil {
		t.Fatalf("validating SSH configuration should parse retry timeout")
	}
}

func TestValidateParseRetryInterval(t *testing.T) {
	t.Parallel()

	c := &Config{
		Address:           "localhost",
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "30s",
		RetryTimeout:      "60s",
		RetryInterval:     "doh",
		Port:              Port,
	}
	if err := c.Validate(); err == nil {
		t.Fatalf("validating SSH configuration should parse retry interval")
	}
}

func TestValidateParsePrivateKey(t *testing.T) {
	t.Parallel()

	c := &Config{
		Address:           "localhost",
		User:              "root",
		ConnectionTimeout: "30s",
		RetryTimeout:      "60s",
		RetryInterval:     "1s",
		Port:              Port,
		PrivateKey:        "foo",
	}
	if err := c.Validate(); err == nil {
		t.Fatalf("validating SSH configuration should parse private key")
	}
}

func generateRSAPrivateKey(t *testing.T) string {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating key failed: %v", err)
	}

	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	return string(pem.EncodeToMemory(&privBlock))
}

func TestHandleClientLocalRemote(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()

	remoteServer, remoteClient := net.Pipe()

	go handleClient(server, remoteServer)

	fmt.Fprint(client, expectedMessage)

	buf := make([]byte, 1024)

	if _, err := remoteClient.Read(buf); err != nil {
		t.Fatalf("reading data from connection should succeed, got: %v", err)
	}

	if reflect.DeepEqual(string(buf), expectedMessage) {
		t.Fatalf("bad response. expected '%s', got '%s'", expectedMessage, string(buf))
	}
}

func TestHandleClientRemoteLocal(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()

	remoteServer, remoteClient := net.Pipe()

	go handleClient(server, remoteServer)

	fmt.Fprint(remoteClient, expectedMessage)

	buf := make([]byte, 1024)

	if _, err := client.Read(buf); err != nil {
		t.Fatalf("reading data from connection should succeed, got: %v", err)
	}

	if reflect.DeepEqual(string(buf), expectedMessage) {
		t.Fatalf("bad response. expected '%s', got '%s'", expectedMessage, string(buf))
	}
}

func TestHandleClientBiDirectional(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()

	remoteServer, remoteClient := net.Pipe()

	go handleClient(server, remoteServer)

	fmt.Fprint(client, expectedMessage)

	buf := make([]byte, 1024)

	if _, err := remoteClient.Read(buf); err != nil {
		t.Fatalf("reading data from connection should succeed, got: %v", err)
	}

	if reflect.DeepEqual(string(buf), expectedMessage) {
		t.Fatalf("bad response. expected '%s', got '%s'", expectedMessage, string(buf))
	}

	fmt.Fprint(remoteClient, expectedResponse)

	buf = make([]byte, 1024)

	if _, err := client.Read(buf); err != nil {
		t.Fatalf("reading data from connection should succeed, got: %v", err)
	}

	if reflect.DeepEqual(string(buf), expectedResponse) {
		t.Fatalf("bad response. expected '%s', got '%s'", expectedResponse, string(buf))
	}
}

func TestExtractPath(t *testing.T) {
	t.Parallel()

	expectedPath := "/tmp/foo.sock"

	p, err := extractPath(fmt.Sprintf("unix://%s", expectedPath))
	if err != nil {
		t.Fatalf("extracting valid path should succeed, got: %v", err)
	}

	if p != expectedPath {
		t.Fatalf("expected %s, got %s", expectedPath, p)
	}
}

func TestExtractPathMalformed(t *testing.T) {
	t.Parallel()

	if _, err := extractPath("ddd\t"); err == nil {
		t.Fatalf("extracting malformed path should fail")
	}
}

func TestExtractPathTCP(t *testing.T) {
	t.Parallel()

	if _, err := extractPath("tcp://localhost:25"); err == nil {
		t.Fatalf("extracting path with unsupported scheme should fail")
	}
}

// randomUnixSocket() tests.
func TestRandomUnixSocket(t *testing.T) {
	t.Parallel()

	d := newConnected("localhost:80", nil).(*sshConnected)

	unixAddr, err := d.randomUnixSocket()
	if err != nil {
		t.Fatalf("creating random unix socket shouldn't fail, got: %v", err)
	}

	if !strings.Contains(unixAddr.String(), d.address) {
		t.Fatalf("generated UNIX address should contain original address %s, got: %s", d.address, unixAddr.String())
	}

	if unixAddr.Net != "unix" {
		t.Fatalf("generated UNIX address should be UNIX address, got net %s", unixAddr.Net)
	}
}

func TestRandomUnixSocketBadUUID(t *testing.T) {
	t.Parallel()

	d := newConnected("localhost:80", nil).(*sshConnected)
	d.uuid = func() (uuid.UUID, error) {
		return uuid.UUID{}, fmt.Errorf("happened")
	}

	if _, err := d.randomUnixSocket(); err == nil {
		t.Fatalf("Creating random unix socket should fail")
	}
}

// forwardConnection() tests.
func TestForwardConnection(t *testing.T) {
	t.Parallel()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("unable to listen on random TCP port: %v", err)
	}

	r, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("unable to listen on random TCP port: %v", err)
	}

	go forwardConnection(l, &net.Dialer{}, r.Addr().String(), "tcp")

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("failed opening connection to listener: %v", err)
	}

	data := []byte("FOO")

	if _, err := conn.Write(data); err != nil {
		t.Fatalf("failed writing to connection: %v", err)
	}

	c, err := r.Accept()
	if err != nil {
		t.Fatalf("failed accepting forwarded connection: %v", err)
	}

	buf := make([]byte, 3)

	if _, err := c.Read(buf); err != nil {
		t.Fatalf("failed reading data from connection: %v", err)
	}

	if !reflect.DeepEqual(buf, data) {
		t.Fatalf("Expected data to be '%s', got '%s'", string(data), string(buf))
	}
}

func TestForwardConnectionBadType(t *testing.T) {
	t.Parallel()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("unable to listen on random TCP port: %v", err)
	}

	r, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("unable to listen on random TCP port: %v", err)
	}

	go forwardConnection(l, &net.Dialer{}, r.Addr().String(), "doh")

	// Try to open connection, so forwarding loop breaks.
	if _, err := net.Dial("tcp", l.Addr().String()); err != nil {
		t.Logf("Opening first connection should succeed, got: %v", err)
	}

	time.Sleep(time.Second)

	if _, err := net.Dial("tcp", l.Addr().String()); err == nil {
		t.Fatalf("Opening connection to bad type should fail")
	}
}

func TestForwardConnectionClosedListener(t *testing.T) {
	t.Parallel()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("unable to listen on random TCP port: %v", err)
	}

	if err := l.Close(); err != nil {
		t.Logf("failed to close listener: %v", err)
	}

	r, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("unable to listen on random TCP port: %v", err)
	}

	go forwardConnection(l, &net.Dialer{}, r.Addr().String(), "tcp")

	if _, err := net.Dial("tcp", l.Addr().String()); err == nil {
		t.Fatalf("Opening connection to closed listener should fail")
	}
}

// Connect() tests.
//
// This test may access SSHAuthSockEnv environment variable,
// which is global variable, so to keep things stable, don't run it in parallel.
//
//nolint:paralleltest
func TestConnect(t *testing.T) {
	if err := os.Unsetenv(SSHAuthSockEnv); err != nil {
		t.Fatalf("failed unsetting environment variable %q: %v", SSHAuthSockEnv, err)
	}

	c := &Config{
		Address:           "localhost",
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "30s",
		RetryTimeout:      "60s",
		RetryInterval:     "1s",
		Port:              Port,
		PrivateKey:        generateRSAPrivateKey(t),
	}

	s, err := c.New()
	if err != nil {
		t.Fatalf("creating new SSH object should succeed, got: %s", err)
	}

	ss := s.(*ssh)

	ss.sshClientGetter = func(n, a string, config *gossh.ClientConfig) (*gossh.Client, error) {
		return nil, nil
	}

	if _, err := ss.Connect(); err != nil {
		t.Fatalf("Connecting should succeed, got: %v", err)
	}
}

// This test may access SSHAuthSockEnv environment variable,
// which is global variable, so to keep things stable, don't run it in parallel.
//
//nolint:paralleltest
func TestConnectFail(t *testing.T) {
	if err := os.Unsetenv(SSHAuthSockEnv); err != nil {
		t.Fatalf("failed unsetting environment variable %q: %v", SSHAuthSockEnv, err)
	}

	c := &Config{
		Address:           "localhost",
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "1s",
		RetryTimeout:      "1s",
		RetryInterval:     "1s",
		Port:              Port,
		PrivateKey:        generateRSAPrivateKey(t),
	}

	s, err := c.New()
	if err != nil {
		t.Fatalf("creating new SSH object should succeed, got: %s", err)
	}

	ss := s.(*ssh)

	ss.sshClientGetter = func(n, a string, config *gossh.ClientConfig) (*gossh.Client, error) {
		return nil, fmt.Errorf("expected")
	}

	if _, err := ss.Connect(); err == nil {
		t.Fatalf("Connecting should fail")
	}
}

// ForwardTCP() tests.
func TestForwardTCP(t *testing.T) {
	t.Parallel()

	d := newConnected("localhost:80", nil).(*sshConnected)

	d.listener = func(n, a string) (net.Listener, error) {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("unable to listen on random TCP port: %v", err)
		}

		return l, nil
	}

	if _, err := d.ForwardTCP("localhost:90"); err != nil {
		t.Fatalf("Forwarding TCP shouldn't fail, got: %v", err)
	}
}

func TestForwardTCPFailListen(t *testing.T) {
	t.Parallel()

	d := newConnected("localhost:80", nil).(*sshConnected)

	d.listener = func(n, a string) (net.Listener, error) {
		return nil, fmt.Errorf("expected")
	}

	if _, err := d.ForwardTCP("localhost:90"); err == nil {
		t.Fatalf("Forwarding TCP should fail")
	}
}

func TestForwardTCPValidateAddress(t *testing.T) {
	t.Parallel()

	d := newConnected("localhost:80", nil).(*sshConnected)

	d.listener = func(n, a string) (net.Listener, error) {
		return nil, fmt.Errorf("expected")
	}

	if _, err := d.ForwardTCP("localhost"); err == nil {
		t.Fatalf("Forwarding TCP should fail when forwarding bad address")
	}
}

// ForwardUnixSocket() tests.
func TestForwardUnixSocketNoRandomUnixSocket(t *testing.T) {
	t.Parallel()

	d := newConnected("localhost:80", nil).(*sshConnected)

	d.uuid = func() (uuid.UUID, error) {
		return uuid.UUID{}, fmt.Errorf("happened")
	}

	if _, err := d.ForwardUnixSocket("foo"); err == nil {
		t.Fatalf("Forwarding with bad unix socket should fail")
	}
}

func TestForwardUnixSocketCantListen(t *testing.T) {
	t.Parallel()

	d := newConnected("localhost:80", nil).(*sshConnected)

	d.listener = func(n, a string) (net.Listener, error) {
		return nil, fmt.Errorf("expected")
	}

	if _, err := d.ForwardUnixSocket("foo"); err == nil {
		t.Fatalf("Forwarding with failed listening should fail")
	}
}

func TestForwardUnixSocketBadPath(t *testing.T) {
	t.Parallel()

	d := newConnected("localhost:80", nil).(*sshConnected)

	if _, err := d.ForwardUnixSocket("foo\t"); err == nil {
		t.Fatalf("Forwarding with invalid unix socket name should fail")
	}
}

func TestForwardUnixSocket(t *testing.T) {
	t.Parallel()

	d := newConnected("localhost:80", nil).(*sshConnected)

	if _, err := d.ForwardUnixSocket("unix:///foo"); err != nil {
		t.Fatalf("Forwarding should succeed, got: %v", err)
	}
}

func TestForwardUnixSocketEnsureUnique(t *testing.T) {
	t.Parallel()

	d := newConnected("localhost:80", nil).(*sshConnected)

	a, err := d.ForwardUnixSocket("unix:///foo")
	if err != nil {
		t.Fatalf("forwarding unix socket should succeed, got: %v", err)
	}

	b, err := d.ForwardUnixSocket("unix:///foo")
	if err != nil {
		t.Fatalf("forwarding 2nd random unix socket should succeed, got: %v", err)
	}

	if diff := cmp.Diff(a, b); diff == "" {
		t.Fatalf("forwarded random unix sockets should differ")
	}
}

// This test may access SSHAuthSockEnv environment variable,
// which is global variable, so to keep things stable, don't run it in parallel.
//
//nolint:paralleltest
func TestNewBadSSHAgentEnv(t *testing.T) {
	if err := os.Setenv(SSHAuthSockEnv, "foo"); err != nil {
		t.Fatalf("failed setting environment variable %q: %v", SSHAuthSockEnv, err)
	}

	c := &Config{
		Address:           "localhost",
		User:              "root",
		ConnectionTimeout: "30s",
		RetryTimeout:      "60s",
		RetryInterval:     "1s",
		Port:              Port,
	}

	if _, err := c.New(); err == nil {
		t.Fatalf("creating new SSH object with bad ssh-agent environment variable should fail")
	}
}

func TestNewSSHAgent(t *testing.T) {
	t.Parallel()

	a := agent.NewKeyring()

	addr := &net.UnixAddr{
		Name: "@foo",
		Net:  "unix",
	}

	l, err := net.Listen("unix", addr.String())
	if err != nil {
		t.Fatalf("failed to listen on address %q: %v", addr.String(), err)
	}

	go func() {
		c, err := l.Accept()
		if err != nil {
			t.Logf("accepting connection failed: %v", err)
		}

		if err := agent.ServeAgent(a, c); err != nil {
			t.Logf("serving agent failed: %v", err)
		}

		if err := l.Close(); err != nil {
			t.Logf("closing listener failed: %v", err)
		}
	}()

	if err := os.Setenv(SSHAuthSockEnv, addr.String()); err != nil {
		t.Fatalf("failed setting environment variable %q: %v", SSHAuthSockEnv, err)
	}

	c := &Config{
		Address:           "localhost",
		User:              "root",
		ConnectionTimeout: "30s",
		RetryTimeout:      "60s",
		RetryInterval:     "1s",
		Port:              Port,
	}

	if _, err := c.New(); err != nil {
		t.Fatalf("creating new SSH object with good ssh-agent should work, got: %v", err)
	}
}

func TestNewSSHAgentWrongSocket(t *testing.T) {
	t.Parallel()

	addr := &net.UnixAddr{
		Name: "@bar",
		Net:  "unix",
	}

	l, err := net.Listen("unix", addr.String())
	if err != nil {
		t.Fatalf("failed to listen on address %q: %v", addr.String(), err)
	}

	go func() {
		c, err := l.Accept()
		if err != nil {
			t.Logf("accepting connection failed: %v", err)
		}

		if err := c.Close(); err != nil {
			t.Logf("closing connection failed: %v", err)
		}
	}()

	if err := os.Setenv(SSHAuthSockEnv, addr.String()); err != nil {
		t.Fatalf("failed setting environment variable %q: %v", SSHAuthSockEnv, err)
	}

	c := &Config{
		Address:           "localhost",
		User:              "root",
		ConnectionTimeout: "30s",
		RetryTimeout:      "60s",
		RetryInterval:     "1s",
		Port:              Port,
	}

	if _, err := c.New(); err == nil {
		t.Fatalf("creating new SSH object with bad ssh-agent socket should fail")
	}
}
