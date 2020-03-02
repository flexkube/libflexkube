package flexkube

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
)

func TestSSHMarshal(t *testing.T) {
	c := ssh.Config{
		Address:           "127.0.0.1",
		Port:              examplePort,
		User:              "foo",
		Password:          "bar",
		ConnectionTimeout: "1s",
		RetryTimeout:      "2s",
		RetryInterval:     "3s",
		PrivateKey:        "doh",
	}

	e := []interface{}{
		map[string]interface{}{
			"address":            "127.0.0.1",
			"port":               examplePort,
			"user":               "foo",
			"password":           "bar",
			"connection_timeout": "1s",
			"retry_timeout":      "2s",
			"retry_interval":     "3s",
			"private_key":        "doh",
		},
	}

	if diff := cmp.Diff(sshMarshal(c, false), e); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestSSHMarshalSensitive(t *testing.T) {
	c := ssh.Config{
		Address:           "127.0.0.1",
		Port:              examplePort,
		User:              "foo",
		Password:          "bar",
		ConnectionTimeout: "1s",
		RetryTimeout:      "2s",
		RetryInterval:     "3s",
		PrivateKey:        "doh",
	}

	e := []interface{}{
		map[string]interface{}{
			"address":            "127.0.0.1",
			"port":               examplePort,
			"user":               "foo",
			"password":           sha256sum([]byte("bar")),
			"connection_timeout": "1s",
			"retry_timeout":      "2s",
			"retry_interval":     "3s",
			"private_key":        sha256sum([]byte("doh")),
		},
	}

	if diff := cmp.Diff(sshMarshal(c, true), e); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestSSHMarshalSensitivePassword(t *testing.T) {
	c := ssh.Config{
		Address:           "127.0.0.1",
		Port:              examplePort,
		User:              "foo",
		Password:          "bar",
		ConnectionTimeout: "1s",
		RetryTimeout:      "2s",
		RetryInterval:     "3s",
		PrivateKey:        "",
	}

	e := []interface{}{
		map[string]interface{}{
			"address":            "127.0.0.1",
			"port":               examplePort,
			"user":               "foo",
			"password":           sha256sum([]byte("bar")),
			"connection_timeout": "1s",
			"retry_timeout":      "2s",
			"retry_interval":     "3s",
			"private_key":        "",
		},
	}

	if diff := cmp.Diff(sshMarshal(c, true), e); diff != "" {
		t.Errorf("When private key is not set, it should not be hidden from the state, got diff:\n%s", diff)
	}
}

func TestSSHMarshalSensitivePrivateKey(t *testing.T) {
	c := ssh.Config{
		Address:           "127.0.0.1",
		Port:              examplePort,
		User:              "foo",
		Password:          "",
		ConnectionTimeout: "1s",
		RetryTimeout:      "2s",
		RetryInterval:     "3s",
		PrivateKey:        "doh",
	}

	e := []interface{}{
		map[string]interface{}{
			"address":            "127.0.0.1",
			"port":               examplePort,
			"user":               "foo",
			"password":           "",
			"connection_timeout": "1s",
			"retry_timeout":      "2s",
			"retry_interval":     "3s",
			"private_key":        sha256sum([]byte("doh")),
		},
	}

	if diff := cmp.Diff(sshMarshal(c, true), e); diff != "" {
		t.Errorf("When password is not set, it should not be hidden from the state, got diff:\n%s", diff)
	}
}

func TestSSHUnmarshal(t *testing.T) {
	c := ssh.Config{
		Address:           "127.0.0.1",
		Port:              examplePort,
		User:              "foo",
		Password:          "bar",
		ConnectionTimeout: "1s",
		RetryTimeout:      "2s",
		RetryInterval:     "3s",
		PrivateKey:        "doh",
	}

	e := []interface{}{
		map[string]interface{}{
			"address":            "127.0.0.1",
			"port":               examplePort,
			"user":               "foo",
			"password":           "bar",
			"connection_timeout": "1s",
			"retry_timeout":      "2s",
			"retry_interval":     "3s",
			"private_key":        "doh",
		},
	}

	if diff := cmp.Diff(sshUnmarshal(e[0]), &c); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestSSHUnmarshalEmpty(t *testing.T) {
	var c *ssh.Config

	if diff := cmp.Diff(sshUnmarshal(nil), c); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestSSHUnmarshalEmptyBock(t *testing.T) {
	var c *ssh.Config

	if diff := cmp.Diff(sshUnmarshal(map[string]interface{}{}), c); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}
