package ssh

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/rand"
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
	authMethods = 1
)

func unsetSSHAuthSockEnv(t *testing.T) {
	t.Helper()

	if err := os.Unsetenv(SSHAuthSockEnv); err != nil {
		t.Fatalf("failed unsetting environment variable %q: %v", SSHAuthSockEnv, err)
	}
}

//nolint:paralleltest // This test may access SSHAuthSockEnv environment variable,
// which is a global variable, so to keep things stable, don't run it in parallel.
func TestNew(t *testing.T) {
	unsetSSHAuthSockEnv(t)

	c := testConfig(t)

	if _, err := c.New(); err != nil {
		t.Fatalf("creating new SSH object should succeed, got: %s", err)
	}
}

//nolint:paralleltest // This test may access SSHAuthSockEnv environment variable,
// which is a global variable, so to keep things stable, don't run it in parallel.
func TestNewSetPassword(t *testing.T) {
	unsetSSHAuthSockEnv(t)

	c := testConfig(t)
	c.PrivateKey = ""

	s, err := c.New()
	if err != nil {
		t.Fatalf("creating new SSH object should succeed, got: %s", err)
	}

	if len(s.(*ssh).auth) != authMethods {
		t.Fatalf("when Password field is set, object should include one auth method")
	}
}

//nolint:paralleltest // This test may access SSHAuthSockEnv environment variable,
// which is a global variable, so to keep things stable, don't run it in parallel.
func TestNewSetPrivateKey(t *testing.T) {
	unsetSSHAuthSockEnv(t)

	c := testConfig(t)
	c.Password = ""

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

//nolint:paralleltest // This test may access SSHAuthSockEnv environment variable,
// which is a global variable, so to keep things stable, don't run it in parallel.
func TestValidateRequireAuth(t *testing.T) {
	unsetSSHAuthSockEnv(t)

	c := testConfig(t)
	c.PrivateKey = ""
	c.Password = ""

	if err := c.Validate(); err == nil {
		t.Fatalf("validating SSH configuration should require retry interval field")
	}
}

func Test_Validating_config_returns_error_when(t *testing.T) {
	t.Parallel()

	for name, mutateF := range map[string]func(*Config){
		"address_is_empty":                             func(c *Config) { c.Address = "" },
		"user_is_empty":                                func(c *Config) { c.User = "" },
		"connection_timeout_is_empty":                  func(c *Config) { c.ConnectionTimeout = "" },
		"retry_timeout_is_empty":                       func(c *Config) { c.RetryTimeout = "" },
		"retry_interval_is_empty":                      func(c *Config) { c.RetryInterval = "" },
		"port_is_zero":                                 func(c *Config) { c.Port = 0 },
		"connection_timeout_is_not_a_valid_duration":   func(c *Config) { c.ConnectionTimeout = "baz" },
		"retry_timeout_is_not_a_valid_duration":        func(c *Config) { c.RetryTimeout = "bar" },
		"retry_interval_is_not_a_valid_duration":       func(c *Config) { c.RetryInterval = "ban" },
		"private_key_is_not_a_PEM_encoded_private_key": func(c *Config) { c.PrivateKey = "bah" },
	} {
		mutateF := mutateF

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			c := testConfig(t)
			mutateF(c)

			if err := c.Validate(); err == nil {
				t.Fatal("Expected validation error")
			}
		})
	}
}

func testConfig(t *testing.T) *Config {
	t.Helper()

	return &Config{
		Address:           "localhost",
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "30s",
		RetryTimeout:      "60s",
		RetryInterval:     "1s",
		Port:              Port,
		PrivateKey:        generateRSAPrivateKey(t),
	}
}

func generateRSAPrivateKey(t *testing.T) string {
	t.Helper()

	privateKey, err := rsa.GenerateKey(cryptorand.Reader, 2048)
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

const maxTestMessageLength = 1024

func testMessage(t *testing.T) ([]byte, int) {
	t.Helper()

	rand.Seed(time.Now().UTC().UnixNano())

	// We must have at least 1 byte message.
	length := rand.Intn(maxTestMessageLength) + 1

	message := make([]byte, length)
	if _, err := rand.Read(message); err != nil {
		t.Fatalf("generating message: %v", err)
	}

	message = bytes.Trim(message, "\x00")

	return message, len(message)
}

func TestHandleClientLocalRemote(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()

	remoteServer, remoteClient := net.Pipe()

	go handleClient(server, remoteServer)

	expectedMessage, _ := testMessage(t)

	if _, err := client.Write(expectedMessage); err != nil {
		t.Fatalf("Writing to local client failed: %v", err)
	}

	if err := client.Close(); err != nil {
		t.Fatalf("Closing local client failed: %v", err)
	}

	readMessage, err := ioutil.ReadAll(remoteClient)
	if err != nil {
		t.Fatalf("reading data from connection should succeed, got: %v", err)
	}

	if !reflect.DeepEqual(readMessage, expectedMessage) {
		t.Fatalf("bad response. expected '%+v', got '%+v'", expectedMessage, readMessage)
	}
}

func TestHandleClientRemoteLocal(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()

	remoteServer, remoteClient := net.Pipe()

	go handleClient(server, remoteServer)

	expectedMessage, _ := testMessage(t)

	if _, err := remoteClient.Write(expectedMessage); err != nil {
		t.Fatalf("Writing to remote client failed: %v", err)
	}

	if err := remoteClient.Close(); err != nil {
		t.Fatalf("Closing remote client failed: %v", err)
	}

	readMessage, err := ioutil.ReadAll(client)
	if err != nil {
		t.Fatalf("reading data from connection should succeed, got: %v", err)
	}

	if !reflect.DeepEqual(readMessage, expectedMessage) {
		t.Fatalf("bad response. expected:\n '%+v'\n got:\n '%+v'", expectedMessage, readMessage)
	}
}

func TestHandleClientBiDirectional(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()

	remoteServer, remoteClient := net.Pipe()

	go handleClient(server, remoteServer)

	randomRequest, requestLength := testMessage(t)

	if _, err := client.Write(randomRequest); err != nil {
		t.Fatalf("Writing to local client failed: %v", err)
	}

	// Read twice as much data as we send to make sure we don't send any extra garbage.
	receivedRequest := make([]byte, requestLength*2)

	bytesRead, err := remoteClient.Read(receivedRequest)
	if err != nil {
		t.Fatalf("Reading data from connection should succeed, got: %v", err)
	}

	if bytesRead != requestLength {
		t.Fatalf("%d differs from %d", bytesRead, requestLength)
	}

	// Get rid of any extra null bytes before comparison, as we have more in slice than we read.
	receivedRequest = bytes.TrimRight(receivedRequest, "\x00")

	if !reflect.DeepEqual(randomRequest, receivedRequest) {
		t.Fatalf("Bad response. expected '%+v', got '%+v'", randomRequest, receivedRequest)
	}

	randomResponse, _ := testMessage(t)

	if _, err := remoteClient.Write(randomResponse); err != nil {
		t.Fatalf("Writing to remote client failed: %v", err)
	}

	if err := remoteClient.Close(); err != nil {
		t.Fatalf("Closing remote client failed: %v", err)
	}

	receivedResponse, err := ioutil.ReadAll(client)
	if err != nil {
		t.Fatalf("Reading data from connection should succeed, got: %v", err)
	}

	if !reflect.DeepEqual(randomResponse, receivedResponse) {
		t.Fatalf("Bad response. expected '%+v', got '%+v'", randomResponse, receivedResponse)
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

func testNewConnected(t *testing.T) *sshConnected {
	t.Helper()

	c, ok := newConnected("localhost:80", nil).(*sshConnected)
	if !ok {
		t.Fatalf("converting connected to internal state")
	}

	return c
}

// randomUnixSocket() tests.
func TestRandomUnixSocket(t *testing.T) {
	t.Parallel()

	d := testNewConnected(t)

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

	d := testNewConnected(t)
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

	randomRequest, _ := testMessage(t)

	if _, err := conn.Write(randomRequest); err != nil {
		t.Fatalf("failed writing to connection: %v", err)
	}

	// Close connection so we can use ReadAll().
	if err := conn.Close(); err != nil {
		t.Fatalf("Closing connection failed: %v", err)
	}

	c, err := r.Accept()
	if err != nil {
		t.Fatalf("failed accepting forwarded connection: %v", err)
	}

	readData, err := ioutil.ReadAll(c)
	if err != nil {
		t.Fatalf("failed reading data from connection: %v", err)
	}

	if !reflect.DeepEqual(readData, randomRequest) {
		t.Fatalf("Expected data to be '%+v', got '%+v'", randomRequest, readData)
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
//nolint:paralleltest // This test may access SSHAuthSockEnv environment variable,
// which is a global variable, so to keep things stable, don't run it in parallel.
func TestConnect(t *testing.T) {
	unsetSSHAuthSockEnv(t)

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

	ss, ok := s.(*ssh)
	if !ok {
		t.Fatalf("Converting SSH to internal SSH object")
	}

	ss.sshClientGetter = func(n, a string, config *gossh.ClientConfig) (*gossh.Client, error) {
		return nil, nil
	}

	if _, err := ss.Connect(); err != nil {
		t.Fatalf("Connecting should succeed, got: %v", err)
	}
}

//nolint:paralleltest // This test may access SSHAuthSockEnv environment variable,
// which is a global variable, so to keep things stable, don't run it in parallel.
func TestConnectFail(t *testing.T) {
	unsetSSHAuthSockEnv(t)

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

	ss, ok := s.(*ssh)
	if !ok {
		t.Fatalf("Converting SSH to internal SSH object")
	}

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

	d := testNewConnected(t)

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

	d := testNewConnected(t)

	d.listener = func(n, a string) (net.Listener, error) {
		return nil, fmt.Errorf("expected")
	}

	if _, err := d.ForwardTCP("localhost:90"); err == nil {
		t.Fatalf("Forwarding TCP should fail")
	}
}

func TestForwardTCPValidateAddress(t *testing.T) {
	t.Parallel()

	d := testNewConnected(t)

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

	d := testNewConnected(t)

	d.uuid = func() (uuid.UUID, error) {
		return uuid.UUID{}, fmt.Errorf("happened")
	}

	if _, err := d.ForwardUnixSocket("foo"); err == nil {
		t.Fatalf("Forwarding with bad unix socket should fail")
	}
}

func TestForwardUnixSocketCantListen(t *testing.T) {
	t.Parallel()

	d := testNewConnected(t)

	d.listener = func(n, a string) (net.Listener, error) {
		return nil, fmt.Errorf("expected")
	}

	if _, err := d.ForwardUnixSocket("foo"); err == nil {
		t.Fatalf("Forwarding with failed listening should fail")
	}
}

func TestForwardUnixSocketBadPath(t *testing.T) {
	t.Parallel()

	d := testNewConnected(t)

	if _, err := d.ForwardUnixSocket("foo\t"); err == nil {
		t.Fatalf("Forwarding with invalid unix socket name should fail")
	}
}

func TestForwardUnixSocket(t *testing.T) {
	t.Parallel()

	d := testNewConnected(t)

	if _, err := d.ForwardUnixSocket("unix:///foo"); err != nil {
		t.Fatalf("Forwarding should succeed, got: %v", err)
	}
}

func TestForwardUnixSocketEnsureUnique(t *testing.T) {
	t.Parallel()

	d := testNewConnected(t)

	a, err := d.ForwardUnixSocket("unix:///foo")
	if err != nil {
		t.Fatalf("Forwarding unix socket should succeed, got: %v", err)
	}

	b, err := d.ForwardUnixSocket("unix:///foo")
	if err != nil {
		t.Fatalf("Forwarding 2nd random unix socket should succeed, got: %v", err)
	}

	if diff := cmp.Diff(a, b); diff == "" {
		t.Fatalf("Forwarded random unix sockets should differ")
	}
}

//nolint:paralleltest // This test may access SSHAuthSockEnv environment variable,
// which is a global variable, so to keep things stable, don't run it in parallel.
func TestNewBadSSHAgentEnv(t *testing.T) {
	unsetSSHAuthSockEnv(t)

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
			fmt.Printf("Accepting connection failed: %v\n", err)
		}

		if err := agent.ServeAgent(a, c); err != nil {
			fmt.Printf("Serving agent failed: %v\n", err)
		}

		if err := l.Close(); err != nil {
			fmt.Printf("Closing listener failed: %v\n", err)
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
