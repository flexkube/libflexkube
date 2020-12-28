// +build integration

package ssh

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/flexkube/libflexkube/pkg/host/transport"
)

// This socket must be created on SSH target host (so when running tests in container,
// share /run with the host).
const testServerAddr = "/run/test.sock"

func TestPasswordAuth(t *testing.T) {
	t.Parallel()

	pass, err := ioutil.ReadFile("/home/core/.password")
	if err != nil {
		t.Fatalf("reading password shouldn't fail, got: %v", err)
	}

	c := &Config{
		Address:           "localhost",
		User:              "core",
		ConnectionTimeout: "5s",
		RetryTimeout:      "5s",
		RetryInterval:     "1s",
		Port:              Port,
		Password:          strings.TrimSpace(string(pass)),
	}

	s, err := c.New()
	if err != nil {
		t.Fatalf("creating new SSH object should succeed, got: %v", err)
	}

	if _, err := s.Connect(); err != nil {
		t.Fatalf("connecting should succeed, got: %v", err)
	}
}

func TestPasswordAuthFail(t *testing.T) {
	t.Parallel()

	c := &Config{
		Address:           "localhost",
		User:              "core",
		ConnectionTimeout: "5s",
		RetryTimeout:      "5s",
		RetryInterval:     "1s",
		Port:              Port,
		Password:          "badpassword",
	}

	s, err := c.New()
	if err != nil {
		t.Fatalf("creating new SSH object should succeed, got: %v", err)
	}

	if _, err := s.Connect(); err == nil {
		t.Fatalf("connecting with bad password should fail")
	}
}

func TestPrivateKeyAuth(t *testing.T) {
	t.Parallel()

	s := withPrivateKey(t)

	if _, err := s.Connect(); err != nil {
		t.Fatalf("connecting should succeed, got: %v", err)
	}
}

func withPrivateKey(t *testing.T) transport.Interface {
	t.Helper()
	t.Parallel()

	key, err := ioutil.ReadFile("/home/core/.ssh/id_rsa")
	if err != nil {
		t.Fatalf("reading SSH private key shouldn't fail, got: %v", err)
	}

	c := &Config{
		Address:           "localhost",
		User:              "core",
		ConnectionTimeout: "5s",
		RetryTimeout:      "5s",
		RetryInterval:     "1s",
		Port:              Port,
		PrivateKey:        string(key),
	}

	ssh, err := c.New()
	if err != nil {
		t.Fatalf("creating new SSH object should succeed, got: %v", err)
	}

	return ssh
}

func TestForwardUnixSocketFull(t *testing.T) {
	t.Parallel()

	ssh := withPrivateKey(t)
	expectedMessage := "foo"
	expectedResponse := "bar"

	c, err := ssh.Connect()
	if err != nil {
		t.Fatalf("Connecting should succeed, got: %v", err)
	}

	s, err := c.ForwardUnixSocket(fmt.Sprintf("unix://%s", testServerAddr))
	if err != nil {
		t.Fatalf("forwarding should succeed, got: %v", err)
	}

	go runServer(t, expectedMessage, expectedResponse)

	conn, err := net.Dial("unix", strings.ReplaceAll(s, "unix://", ""))
	if err != nil {
		t.Fatalf("opening connection to %s should succeed, got: %v", s, err)
	}

	fmt.Fprint(conn, expectedMessage)

	buf := make([]byte, 1024)

	_, err = conn.Read(buf)
	if err != nil {
		t.Fatalf("reading data from connection should succeed, got: %v", err)
	}

	if reflect.DeepEqual(string(buf), expectedResponse) {
		t.Fatalf("bad response. expected '%s', got '%s'", expectedResponse, string(buf))
	}
}

func runServer(t *testing.T, expectedMessage, response string) {
	t.Helper()

	l, err := net.Listen("unix", testServerAddr)
	if err != nil {
		// Can't use t.Fatalf from go routine. use fmt.Printf + t.Fail() instead
		//
		// SA2002: the goroutine calls T.Fatalf, which must be called in the same goroutine as the test (staticcheck)
		fmt.Printf("listening on socket should succeed, got: %v\n", err)
		t.Fail()
	}

	// We may SSH into host as unprivileged user, so make sure we are allowed to access the
	// socket file.
	if err := os.Chmod(testServerAddr, 0o600); err != nil {
		fmt.Printf("socket chmod should succeed, got: %v\n", err)
		t.Fail()
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Printf("accepting connection should succeed, got: %v\n", err)
		t.Fail()
	}

	buf := make([]byte, 1024)

	_, err = conn.Read(buf)
	if err != nil {
		fmt.Printf("reading data from connection should succeed, got: %v\n", err)
		t.Fail()
	}

	if reflect.DeepEqual(string(buf), expectedMessage) {
		fmt.Printf("bad message. expected '%s', got '%s'\n", expectedMessage, string(buf))
		t.Fail()
	}

	if _, err := conn.Write([]byte(response)); err != nil {
		fmt.Printf("writing response should succeed, got: %v\n", err)
		t.Fail()
	}

	if err := conn.Close(); err != nil {
		t.Logf("failed closing connection: %v", err)
	}

	if err := l.Close(); err != nil {
		t.Logf("failed closing local listener: %v", err)
	}
}
