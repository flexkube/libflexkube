// +build integration

package ssh

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestPasswordAuth(t *testing.T) {
	pass, err := ioutil.ReadFile("/home/core/.password")
	if err != nil {
		t.Fatalf("reading password shouldn't fail, got: %v", err)
	}

	c := &Config{
		Address:           "localhost",
		User:              "core",
		ConnectionTimeout: "5s",
		Port:              22,
		Password:          strings.TrimSpace(string(pass)),
	}

	ssh, err := c.New()
	if err != nil {
		t.Fatalf("creating new SSH object should succeed, got: %v", err)
	}

	if _, err := ssh.ForwardUnixSocket("unix:///run/docker.sock"); err != nil {
		t.Fatalf("forwarding using password should succeed, got: %v", err)
	}
}

func TestPrivateKeyAuth(t *testing.T) {
	key, err := ioutil.ReadFile("/home/core/.ssh/id_rsa")
	if err != nil {
		t.Fatalf("reading SSH private key shouldn't fail, got: %v", err)
	}

	c := &Config{
		Address:           "localhost",
		User:              "core",
		ConnectionTimeout: "5s",
		Port:              22,
		PrivateKey:        string(key),
	}

	ssh, err := c.New()
	if err != nil {
		t.Fatalf("creating new SSH object should succeed, got: %v", err)
	}

	if _, err := ssh.ForwardUnixSocket("unix:///run/docker.sock"); err != nil {
		t.Fatalf("forwarding using SSH private key should succeed, got: %v", err)
	}
}
