package ssh

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
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
		t.Fatalf("Failed unsetting environment variable %q: %v", SSHAuthSockEnv, err)
	}
}

//nolint:paralleltest // This test may access SSHAuthSockEnv environment variable,
//nolint:paralleltest // which is a global variable, so to keep things stable, don't run it in parallel.
func TestNew(t *testing.T) {
	unsetSSHAuthSockEnv(t)

	testConfig := newTestConfig(t)

	if _, err := testConfig.New(); err != nil {
		t.Fatalf("Creating new SSH object should succeed, got: %s", err)
	}
}

//nolint:paralleltest // This test may access SSHAuthSockEnv environment variable,
//nolint:paralleltest // which is a global variable, so to keep things stable, don't run it in parallel.
func TestNewSetPassword(t *testing.T) {
	unsetSSHAuthSockEnv(t)

	testConfig := newTestConfig(t)
	testConfig.PrivateKey = ""
	testConfig.Dialer = func(_, _ string, config *gossh.ClientConfig) (Dialer, error) {
		if len(config.Auth) != authMethods {
			t.Fatalf("Unexpected auth methods, expected %d, got %v", authMethods, config.Auth)
		}

		return &gossh.Client{}, nil
	}

	s, err := testConfig.New()
	if err != nil {
		t.Fatalf("Creating new SSH object should succeed, got: %s", err)
	}

	if _, err := s.Connect(); err != nil {
		t.Fatalf("Unexpected error connecting with dialer: %v", err)
	}
}

//nolint:paralleltest // This test may access SSHAuthSockEnv environment variable,
//nolint:paralleltest // which is a global variable, so to keep things stable, don't run it in parallel.
func TestNewSetPrivateKey(t *testing.T) {
	unsetSSHAuthSockEnv(t)

	testConfig := newTestConfig(t)
	testConfig.Password = ""
	testConfig.Dialer = func(_, _ string, config *gossh.ClientConfig) (Dialer, error) {
		if len(config.Auth) != authMethods {
			t.Fatalf("Unexpected auth methods, expected %d, got %v", authMethods, config.Auth)
		}

		return &gossh.Client{}, nil
	}

	s, err := testConfig.New()
	if err != nil {
		t.Fatalf("Creating new SSH object should succeed, got: %s", err)
	}

	if _, err := s.Connect(); err != nil {
		t.Fatalf("Unexpected error connecting with dialer: %v", err)
	}
}

func TestNewValidate(t *testing.T) {
	t.Parallel()

	testConfig := &Config{}
	if _, err := testConfig.New(); err == nil {
		t.Fatalf("Creating new SSH object should validate it")
	}
}

//nolint:paralleltest // This test may access SSHAuthSockEnv environment variable,
//nolint:paralleltest // which is a global variable, so to keep things stable, don't run it in parallel.
func TestValidateRequireAuth(t *testing.T) {
	unsetSSHAuthSockEnv(t)

	c := newTestConfig(t)
	c.PrivateKey = ""
	c.Password = ""

	if err := c.Validate(); err == nil {
		t.Fatalf("Validating SSH configuration should require either password or private key")
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

			c := newTestConfig(t)
			mutateF(c)

			if err := c.Validate(); err == nil {
				t.Fatal("Expected validation error")
			}
		})
	}
}

func newTestConfig(t *testing.T) *Config {
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

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Generating key failed: %v", err)
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

	randLength, err := rand.Int(rand.Reader, big.NewInt(maxTestMessageLength))
	if err != nil {
		t.Fatalf("Generating random length: %v", err)
	}

	// We must have at least 1 byte message.
	length := randLength.Int64() + 1

	message := make([]byte, length)
	if _, err := rand.Read(message); err != nil {
		t.Fatalf("Generating message: %v", err)
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

	readMessage, err := io.ReadAll(remoteClient)
	if err != nil {
		t.Fatalf("Reading data from connection should succeed, got: %v", err)
	}

	if !reflect.DeepEqual(readMessage, expectedMessage) {
		t.Fatalf("Bad response. expected '%+v', got '%+v'", expectedMessage, readMessage)
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

	readMessage, err := io.ReadAll(client)
	if err != nil {
		t.Fatalf("Reading data from connection should succeed, got: %v", err)
	}

	if !reflect.DeepEqual(readMessage, expectedMessage) {
		t.Fatalf("Bad response. expected:\n '%+v'\n got:\n '%+v'", expectedMessage, readMessage)
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

	receivedResponse, err := io.ReadAll(client)
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

	path, err := extractPath(fmt.Sprintf("unix://%s", expectedPath))
	if err != nil {
		t.Fatalf("Extracting valid path should succeed, got: %v", err)
	}

	if path != expectedPath {
		t.Fatalf("Expected %s, got %s", expectedPath, path)
	}
}

func TestExtractPathMalformed(t *testing.T) {
	t.Parallel()

	if _, err := extractPath("ddd\t"); err == nil {
		t.Fatalf("Extracting malformed path should fail")
	}
}

func TestExtractPathTCP(t *testing.T) {
	t.Parallel()

	if _, err := extractPath("tcp://localhost:25"); err == nil {
		t.Fatalf("Extracting path with unsupported scheme should fail")
	}
}

func testNewConnected(t *testing.T) *sshConnected {
	t.Helper()

	c, ok := newConnected("localhost:80", nil).(*sshConnected)
	if !ok {
		t.Fatalf("Converting connected to internal state")
	}

	return c
}

// randomUnixSocket() tests.
func TestRandomUnixSocket(t *testing.T) {
	t.Parallel()

	connected := testNewConnected(t)

	unixAddr, err := connected.randomUnixSocket()
	if err != nil {
		t.Fatalf("Creating random unix socket shouldn't fail, got: %v", err)
	}

	if !strings.Contains(unixAddr.String(), connected.address) {
		t.Fatalf("Generated UNIX address should contain original address %s, got: %s", connected.address, unixAddr.String())
	}

	if unixAddr.Net != "unix" {
		t.Fatalf("Generated UNIX address should be UNIX address, got net %s", unixAddr.Net)
	}
}

func TestRandomUnixSocketBadUUID(t *testing.T) {
	t.Parallel()

	connected := testNewConnected(t)
	connected.uuid = func() (uuid.UUID, error) {
		return uuid.UUID{}, fmt.Errorf("happened")
	}

	if _, err := connected.randomUnixSocket(); err == nil {
		t.Fatalf("Creating random unix socket should fail")
	}
}

// forwardConnection() tests.
func TestForwardConnection(t *testing.T) {
	t.Parallel()

	forwardListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Unable to listen on random TCP port: %v", err)
	}

	targetListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Unable to listen on random TCP port: %v", err)
	}

	go forwardConnection(forwardListener, &net.Dialer{}, targetListener.Addr().String(), "tcp")

	conn, err := net.Dial("tcp", forwardListener.Addr().String())
	if err != nil {
		t.Fatalf("Failed opening connection to listener: %v", err)
	}

	randomRequest, _ := testMessage(t)

	if _, err := conn.Write(randomRequest); err != nil {
		t.Fatalf("Failed writing to connection: %v", err)
	}

	// Close connection so we can use ReadAll().
	if err := conn.Close(); err != nil {
		t.Fatalf("Closing connection failed: %v", err)
	}

	c, err := targetListener.Accept()
	if err != nil {
		t.Fatalf("Failed accepting forwarded connection: %v", err)
	}

	readData, err := io.ReadAll(c)
	if err != nil {
		t.Fatalf("Failed reading data from connection: %v", err)
	}

	if !reflect.DeepEqual(readData, randomRequest) {
		t.Fatalf("Expected data to be '%+v', got '%+v'", randomRequest, readData)
	}
}

func TestForwardConnectionBadType(t *testing.T) {
	t.Parallel()

	forwardListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Unable to listen on random TCP port: %v", err)
	}

	r, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Unable to listen on random TCP port: %v", err)
	}

	go forwardConnection(forwardListener, &net.Dialer{}, r.Addr().String(), "doh")

	// Try to open connection, so forwarding loop breaks.
	if _, err := net.Dial("tcp", forwardListener.Addr().String()); err != nil {
		t.Logf("Opening first connection should succeed, got: %v", err)
	}

	time.Sleep(time.Second)

	if _, err := net.Dial("tcp", forwardListener.Addr().String()); err == nil {
		t.Fatalf("Opening connection to bad type should fail")
	}
}

func TestForwardConnectionClosedListener(t *testing.T) {
	t.Parallel()

	forwardListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Unable to listen on random TCP port: %v", err)
	}

	if err := forwardListener.Close(); err != nil {
		t.Logf("Failed to close listener: %v", err)
	}

	r, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Unable to listen on random TCP port: %v", err)
	}

	go forwardConnection(forwardListener, &net.Dialer{}, r.Addr().String(), "tcp")

	if _, err := net.Dial("tcp", forwardListener.Addr().String()); err == nil {
		t.Fatalf("Opening connection to closed listener should fail")
	}
}

// Connect() tests.
//
//nolint:paralleltest // This test may access SSHAuthSockEnv environment variable,
//nolint:paralleltest // which is a global variable, so to keep things stable, don't run it in parallel.
func TestConnect(t *testing.T) {
	unsetSSHAuthSockEnv(t)

	testConfig := &Config{
		Address:           "localhost",
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "30s",
		RetryTimeout:      "60s",
		RetryInterval:     "1s",
		Port:              Port,
		PrivateKey:        generateRSAPrivateKey(t),
		Dialer: func(string, string, *gossh.ClientConfig) (Dialer, error) {
			return &gossh.Client{}, nil
		},
	}

	s, err := testConfig.New()
	if err != nil {
		t.Fatalf("Creating new SSH object should succeed, got: %s", err)
	}

	if _, err := s.Connect(); err != nil {
		t.Fatalf("Connecting should succeed, got: %v", err)
	}
}

//nolint:paralleltest // This test may access SSHAuthSockEnv environment variable,
//nolint:paralleltest // which is a global variable, so to keep things stable, don't run it in parallel.
func TestConnectFail(t *testing.T) {
	unsetSSHAuthSockEnv(t)

	testConfig := &Config{
		Address:           "localhost",
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "1s",
		RetryTimeout:      "1s",
		RetryInterval:     "1s",
		Port:              Port,
		PrivateKey:        generateRSAPrivateKey(t),
		Dialer: func(string, string, *gossh.ClientConfig) (Dialer, error) {
			return nil, fmt.Errorf("expected")
		},
	}

	s, err := testConfig.New()
	if err != nil {
		t.Fatalf("Creating new SSH object should succeed, got: %s", err)
	}

	if _, err := s.Connect(); err == nil {
		t.Fatalf("Connecting should fail")
	}
}

// ForwardTCP() tests.
func TestForwardTCP(t *testing.T) {
	t.Parallel()

	connected := testNewConnected(t)

	connected.listener = func(string, string) (net.Listener, error) {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("Unable to listen on random TCP port: %v", err)
		}

		return l, nil
	}

	if _, err := connected.ForwardTCP("localhost:90"); err != nil {
		t.Fatalf("Forwarding TCP shouldn't fail, got: %v", err)
	}
}

func TestForwardTCPFailListen(t *testing.T) {
	t.Parallel()

	connected := testNewConnected(t)

	connected.listener = func(string, string) (net.Listener, error) {
		return nil, fmt.Errorf("expected")
	}

	if _, err := connected.ForwardTCP("localhost:90"); err == nil {
		t.Fatalf("Forwarding TCP should fail")
	}
}

func TestForwardTCPValidateAddress(t *testing.T) {
	t.Parallel()

	connected := testNewConnected(t)

	connected.listener = func(string, string) (net.Listener, error) {
		return nil, fmt.Errorf("expected")
	}

	if _, err := connected.ForwardTCP("localhost"); err == nil {
		t.Fatalf("Forwarding TCP should fail when forwarding bad address")
	}
}

// ForwardUnixSocket() tests.
func TestForwardUnixSocketNoRandomUnixSocket(t *testing.T) {
	t.Parallel()

	connected := testNewConnected(t)

	connected.uuid = func() (uuid.UUID, error) {
		return uuid.UUID{}, fmt.Errorf("happened")
	}

	if _, err := connected.ForwardUnixSocket("foo"); err == nil {
		t.Fatalf("Forwarding with bad unix socket should fail")
	}
}

func TestForwardUnixSocketCantListen(t *testing.T) {
	t.Parallel()

	connected := testNewConnected(t)

	connected.listener = func(string, string) (net.Listener, error) {
		return nil, fmt.Errorf("expected")
	}

	if _, err := connected.ForwardUnixSocket("foo"); err == nil {
		t.Fatalf("Forwarding with failed listening should fail")
	}
}

func TestForwardUnixSocketBadPath(t *testing.T) {
	t.Parallel()

	connected := testNewConnected(t)

	if _, err := connected.ForwardUnixSocket("foo\t"); err == nil {
		t.Fatalf("Forwarding with invalid unix socket name should fail")
	}
}

func TestForwardUnixSocket(t *testing.T) {
	t.Parallel()

	connected := testNewConnected(t)

	if _, err := connected.ForwardUnixSocket("unix:///foo"); err != nil {
		t.Fatalf("Forwarding should succeed, got: %v", err)
	}
}

func TestForwardUnixSocketEnsureUnique(t *testing.T) {
	t.Parallel()

	connected := testNewConnected(t)

	firstForwardedSocket, err := connected.ForwardUnixSocket("unix:///foo")
	if err != nil {
		t.Fatalf("Forwarding unix socket should succeed, got: %v", err)
	}

	secondForwardedSocket, err := connected.ForwardUnixSocket("unix:///foo")
	if err != nil {
		t.Fatalf("Forwarding 2nd random unix socket should succeed, got: %v", err)
	}

	if diff := cmp.Diff(firstForwardedSocket, secondForwardedSocket); diff == "" {
		t.Fatalf("Forwarded random unix sockets should differ")
	}
}

//nolint:paralleltest // This test may access SSHAuthSockEnv environment variable,
//nolint:paralleltest // which is a global variable, so to keep things stable, don't run it in parallel.
func TestNewBadSSHAgentEnv(t *testing.T) {
	unsetSSHAuthSockEnv(t)

	testConfig := &Config{
		Address:           "localhost",
		User:              "root",
		ConnectionTimeout: "30s",
		RetryTimeout:      "60s",
		RetryInterval:     "1s",
		Port:              Port,
	}

	if _, err := testConfig.New(); err == nil {
		t.Fatalf("Creating new SSH object with bad ssh-agent environment variable should fail")
	}
}

func TestNewSSHAgent(t *testing.T) {
	agentKeyring := agent.NewKeyring()

	addr := &net.UnixAddr{
		Name: "@foo",
		Net:  "unix",
	}

	agentListener, err := net.Listen("unix", addr.String())
	if err != nil {
		t.Fatalf("Failed to listen on address %q: %v", addr.String(), err)
	}

	go func() {
		c, err := agentListener.Accept()
		if err != nil {
			fmt.Printf("Accepting connection failed: %v\n", err)
		}

		if err := agent.ServeAgent(agentKeyring, c); err != nil {
			fmt.Printf("Serving agent failed: %v\n", err)
		}

		if err := agentListener.Close(); err != nil {
			fmt.Printf("Closing listener failed: %v\n", err)
		}
	}()

	t.Setenv(SSHAuthSockEnv, addr.String())

	testConfig := &Config{
		Address:           "localhost",
		User:              "root",
		ConnectionTimeout: "30s",
		RetryTimeout:      "60s",
		RetryInterval:     "1s",
		Port:              Port,
	}

	if _, err := testConfig.New(); err != nil {
		t.Fatalf("Creating new SSH object with good ssh-agent should work, got: %v", err)
	}
}

func TestNewSSHAgentWrongSocket(t *testing.T) {
	addr := &net.UnixAddr{
		Name: "@bar",
		Net:  "unix",
	}

	badAgentListener, err := net.Listen("unix", addr.String())
	if err != nil {
		t.Fatalf("Failed to listen on address %q: %v", addr.String(), err)
	}

	go func() {
		c, err := badAgentListener.Accept()
		if err != nil {
			t.Logf("Accepting connection failed: %v", err)
		}

		if err := c.Close(); err != nil {
			t.Logf("Closing connection failed: %v", err)
		}
	}()

	t.Setenv(SSHAuthSockEnv, addr.String())

	testConfig := &Config{
		Address:           "localhost",
		User:              "root",
		ConnectionTimeout: "30s",
		RetryTimeout:      "60s",
		RetryInterval:     "1s",
		Port:              Port,
	}

	if _, err := testConfig.New(); err == nil {
		t.Fatalf("Creating new SSH object with bad ssh-agent socket should fail")
	}
}
