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

// Default Docker Host URL
const dockerSocket = "unix:///run/docker.sock"

func TestPasswordAuth(t *testing.T) {
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
		Port:              22,
		Password:          strings.TrimSpace(string(pass)),
	}

	ssh, err := c.New()
	if err != nil {
		t.Fatalf("creating new SSH object should succeed, got: %v", err)
	}

	forward(t, ssh, dockerSocket)
}

func TestPrivateKeyAuth(t *testing.T) {
	ssh := withPrivateKey(t)

	forward(t, ssh, dockerSocket)
}

func forward(t *testing.T, ssh transport.Transport, path string) string {
	s, err := ssh.ForwardUnixSocket(path)
	if err != nil {
		t.Fatalf("forwarding should succeed, got: %v", err)
	}

	return s
}

func withPrivateKey(t *testing.T) transport.Transport {
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
		Port:              22,
		PrivateKey:        string(key),
	}

	ssh, err := c.New()
	if err != nil {
		t.Fatalf("creating new SSH object should succeed, got: %v", err)
	}

	return ssh
}

func TestForwardUnixSocket(t *testing.T) {
	ssh := withPrivateKey(t)
	expectedMessage := "foo"
	expectedResponse := "bar"
	s := forward(t, ssh, fmt.Sprintf("unix://%s", testServerAddr))

	go runServer(t, expectedMessage, expectedResponse)

	conn, err := net.Dial("unix", strings.Replace(s, "unix://", "", -1))
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

func runServer(t *testing.T, expectedMessage string, response string) {
	l, err := net.Listen("unix", testServerAddr)
	if err != nil {
		// Can't use t.Fatalf from go routine. use fmt.Printf + t.Fail() instead
		//
		// SA2002: the goroutine calls T.Fatalf, which must be called in the same goroutine as the test (staticcheck)
		fmt.Printf("listening on socket should succeed, got: %v\n", err)
		t.Fail()
	}

	defer l.Close()

	// We may SSH into host as unprivileged user, so make sure we are allowed to access the
	// socket file.
	if err := os.Chmod(testServerAddr, 0777); err != nil {
		fmt.Printf("socket chmod should succeed, got: %v\n", err)
		t.Fail()
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Printf("accepting connection should succeed, got: %v\n", err)
		t.Fail()
	}

	defer conn.Close()

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
}
