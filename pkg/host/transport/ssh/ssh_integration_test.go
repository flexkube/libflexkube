//go:build integration
// +build integration

package ssh

import (
	"bytes"
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

//nolint:paralleltest // This test may access SSHAuthSockEnv environment variable,
// which is a global variable, so to keep things stable, don't run it in parallel.
func TestPasswordAuth(t *testing.T) {
	unsetSSHAuthSockEnv(t)

	passwordFilePath := "/home/core/.password"

	pass, err := ioutil.ReadFile(passwordFilePath)
	if err != nil {
		t.Fatalf("Reading password file %q: %v", passwordFilePath, err)
	}

	testConfig := &Config{
		Address:           "localhost",
		User:              "core",
		ConnectionTimeout: "5s",
		RetryTimeout:      "5s",
		RetryInterval:     "1s",
		Port:              Port,
		Password:          strings.TrimSpace(string(pass)),
	}

	testSSH, err := testConfig.New()
	if err != nil {
		t.Fatalf("Creating new SSH object should succeed, got: %v", err)
	}

	if _, err := testSSH.Connect(); err != nil {
		t.Fatalf("Connecting should succeed, got: %v", err)
	}
}

//nolint:paralleltest // This test may access SSHAuthSockEnv environment variable,
// which is a global variable, so to keep things stable, don't run it in parallel.
func TestPasswordAuthFail(t *testing.T) {
	unsetSSHAuthSockEnv(t)

	testConfig := &Config{
		Address:           "localhost",
		User:              "core",
		ConnectionTimeout: "5s",
		RetryTimeout:      "5s",
		RetryInterval:     "1s",
		Port:              Port,
		Password:          "badpassword",
	}

	testSSH, err := testConfig.New()
	if err != nil {
		t.Fatalf("Creating new SSH object should succeed, got: %v", err)
	}

	if _, err := testSSH.Connect(); err == nil {
		t.Fatalf("Connecting with bad password should fail")
	}
}

//nolint:paralleltest // This test may access SSHAuthSockEnv environment variable,
// which is a global variable, so to keep things stable, don't run it in parallel.
func TestPrivateKeyAuth(t *testing.T) {
	s := withPrivateKey(t)

	if _, err := s.Connect(); err != nil {
		t.Fatalf("Connecting should succeed, got: %v", err)
	}
}

func withPrivateKey(t *testing.T) transport.Interface {
	t.Helper()

	unsetSSHAuthSockEnv(t)

	sshPrivateKeyPath := "/home/core/.ssh/id_rsa"

	key, err := ioutil.ReadFile(sshPrivateKeyPath)
	if err != nil {
		t.Fatalf("Reading SSH private key from %q shouldn't fail, got: %v", sshPrivateKeyPath, err)
	}

	withPrivateKeyConfig := &Config{
		Address:           "localhost",
		User:              "core",
		ConnectionTimeout: "5s",
		RetryTimeout:      "5s",
		RetryInterval:     "1s",
		Port:              Port,
		PrivateKey:        string(key),
	}

	sshWithPrivateKey, err := withPrivateKeyConfig.New()
	if err != nil {
		t.Fatalf("Creating new SSH object should succeed, got: %v", err)
	}

	return sshWithPrivateKey
}

//nolint:paralleltest // This test may access SSHAuthSockEnv environment variable,
// which is a global variable, so to keep things stable, don't run it in parallel.
func TestForwardUnixSocketFull(t *testing.T) {
	ssh := withPrivateKey(t)

	connected, err := ssh.Connect()
	if err != nil {
		t.Fatalf("Connecting should succeed, got: %v", err)
	}

	randomRequest, _ := testMessage(t)
	randomResponse, _ := testMessage(t)

	go runServer(t, randomRequest, randomResponse)

	localSocket, err := connected.ForwardUnixSocket(fmt.Sprintf("unix://%s", testServerAddr))
	if err != nil {
		t.Fatalf("Forwarding should succeed, got: %v", err)
	}

	conn, err := net.Dial("unix", strings.ReplaceAll(localSocket, "unix://", ""))
	if err != nil {
		t.Fatalf("Opening connection to %s should succeed, got: %v", localSocket, err)
	}

	if _, err := conn.Write(randomRequest); err != nil {
		t.Fatalf("Writing data to connection: %v", err)
	}

	response, err := ioutil.ReadAll(conn)
	if err != nil {
		t.Fatalf("Reading data from connection should succeed, got: %v", err)
	}

	if !reflect.DeepEqual(response, randomResponse) {
		t.Fatalf("Bad response. expected '%+v', got '%+v'", randomResponse, response)
	}
}

func prepareTestSocket(t *testing.T) net.Listener {
	t.Helper()

	listener, err := net.Listen("unix", testServerAddr)
	if err != nil {
		// Can't use t.Fatalf from go routine. use fmt.Printf + t.Fail() instead
		//
		// SA2002: the goroutine calls T.Fatalf, which must be called in the same goroutine as the test (staticcheck)
		fmt.Printf("Listening on socket should succeed, got: %v\n", err)
		t.Fail()
	}

	t.Cleanup(func() {
		if err := listener.Close(); err != nil {
			fmt.Printf("Failed closing local listener: %v\n", err)
		}

		if err := os.Remove(testServerAddr); err != nil {
			fmt.Printf("Failed removing UNIX socket %q: %v\n", testServerAddr, err)
		}
	})

	// We may SSH into host as unprivileged user, so make sure we are allowed to access the
	// socket file.
	if err := os.Chmod(testServerAddr, 0o777); err != nil {
		fmt.Printf("Socket chmod should succeed, got: %v\n", err)
		t.Fail()
	}

	return listener
}

//nolint:thelper // This function is actually part of the test.
func runServer(t *testing.T, expectedRequest, response []byte) {
	l := prepareTestSocket(t)

	conn, err := l.Accept()
	if err != nil {
		fmt.Printf("Accepting connection should succeed, got: %v\n", err)
		t.Fail()
	}

	expectedRequestLength := len(expectedRequest)

	receivedRequest := make([]byte, expectedRequestLength*2) //nolint:makezero // We do not append here.

	bytesRead, err := conn.Read(receivedRequest)
	if err != nil {
		fmt.Printf("Reading data from connection should succeed, got: %v\n", err)
		t.Fail()
	}

	if bytesRead != expectedRequestLength {
		fmt.Printf("%d differs from %d\n", bytesRead, expectedRequestLength)
		t.Fail()
	}

	// Get rid of any extra null bytes before comparison, as we have more in slice than we read.
	receivedRequest = bytes.TrimRight(receivedRequest, "\x00")

	if !reflect.DeepEqual(receivedRequest, expectedRequest) {
		fmt.Printf("Bad request. expected '%+v', got '%+v'\n", expectedRequest, receivedRequest)
		t.Fail()
	}

	if _, err := conn.Write(response); err != nil {
		fmt.Printf("Writing response should succeed, got: %v\n", err)
		t.Fail()
	}

	if err := conn.Close(); err != nil {
		fmt.Printf("Failed closing connection: %v\n", err)
		t.Fail()
	}
}
