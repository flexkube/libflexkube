package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"
)

const (
	expectedMessage  = "foo"
	expectedResponse = "bar"
)

func TestNew(t *testing.T) {
	privateKey, err := generateRSAPrivateKey()
	if err != nil {
		t.Fatalf("generating RSA key failed")
	}

	c := &Config{
		Address:           "localhost",
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "30s",
		Port:              22,
		PrivateKey:        privateKey,
	}
	if _, err := c.New(); err != nil {
		t.Fatalf("creating new SSH object should succeed, got: %s", err)
	}
}

func TestNewValidate(t *testing.T) {
	c := &Config{}
	if _, err := c.New(); err == nil {
		t.Fatalf("creating new SSH object should validate it")
	}
}

func TestValidateRequireAddress(t *testing.T) {
	c := &Config{
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "30s",
		Port:              22,
	}
	if _, err := c.New(); err == nil {
		t.Fatalf("validating SSH configuration should require address field")
	}
}

func TestValidateRequireUser(t *testing.T) {
	c := &Config{
		Address:           "localhost",
		Password:          "foo",
		ConnectionTimeout: "30s",
		Port:              22,
	}
	if _, err := c.New(); err == nil {
		t.Fatalf("validating SSH configuration should require user field")
	}
}

func TestValidateRequireAuthMethod(t *testing.T) {
	c := &Config{
		Address:           "localhost",
		User:              "root",
		ConnectionTimeout: "30s",
		Port:              22,
	}
	if _, err := c.New(); err == nil {
		t.Fatalf("validating SSH configuration should require at least one authentication method")
	}
}

func TestValidateRequireConnectionTimeout(t *testing.T) {
	c := &Config{
		Address:  "localhost",
		User:     "root",
		Password: "foo",
		Port:     22,
	}
	if _, err := c.New(); err == nil {
		t.Fatalf("validating SSH configuration should require connection timeout field")
	}
}

func TestValidateRequirePort(t *testing.T) {
	c := &Config{
		Address:           "localhost",
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "30s",
	}
	if _, err := c.New(); err == nil {
		t.Fatalf("validating SSH configuration should require port field")
	}
}

func TestValidateParseConnectionTimeout(t *testing.T) {
	c := &Config{
		Address:           "localhost",
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "30",
		Port:              22,
	}
	if _, err := c.New(); err == nil {
		t.Fatalf("validating SSH configuration should parse connection timeout")
	}
}

func TestValidateParsePrivateKey(t *testing.T) {
	c := &Config{
		Address:           "localhost",
		User:              "root",
		ConnectionTimeout: "30s",
		Port:              22,
		PrivateKey:        "foo",
	}
	if _, err := c.New(); err == nil {
		t.Fatalf("validating SSH configuration should parse private key")
	}
}

func generateRSAPrivateKey() (string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", fmt.Errorf("generating key failed: %w", err)
	}

	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	return string(pem.EncodeToMemory(&privBlock)), nil
}

func TestHandleClientLocalRemote(t *testing.T) {
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
	if _, err := extractPath("ddd"); err == nil {
		t.Fatalf("extracting malformed path should fail")
	}
}

func TestExtractPathTCP(t *testing.T) {
	if _, err := extractPath("tcp://localhost:25"); err == nil {
		t.Fatalf("extracting path with unsupported scheme should fail")
	}
}

func TestRandomUnixSocket(t *testing.T) {
	address := "localhost:80"

	unixAddr, err := randomUnixSocket(address)
	if err != nil {
		t.Fatalf("creating random unix socket shouldn't fail, got: %v", err)
	}

	if !strings.Contains(unixAddr.String(), address) {
		t.Fatalf("generated UNIX address should contain original address %s, got: %s", address, unixAddr.String())
	}

	if unixAddr.Net != "unix" {
		t.Fatalf("generated UNIX address should be UNIX address, got net %s", unixAddr.Net)
	}
}
