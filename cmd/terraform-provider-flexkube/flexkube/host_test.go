package flexkube

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
)

func TestHostMarshal(t *testing.T) {
	c := host.Host{
		DirectConfig: &direct.Config{},
		SSHConfig: &ssh.Config{
			Address:           "127.0.0.1",
			Port:              examplePort,
			User:              "foo",
			Password:          "bar",
			ConnectionTimeout: "1s",
			RetryTimeout:      "2s",
			RetryInterval:     "3s",
			PrivateKey:        "doh",
		},
	}

	e := []interface{}{
		map[string]interface{}{
			"direct": []interface{}{
				map[string]interface{}{},
			},
			"ssh": []interface{}{
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
			},
		},
	}

	if diff := cmp.Diff(hostMarshal(c, false), e); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestHostMarshalSensitive(t *testing.T) {
	c := host.Host{
		DirectConfig: &direct.Config{},
		SSHConfig: &ssh.Config{
			Address:           "127.0.0.1",
			Port:              examplePort,
			User:              "foo",
			Password:          "bar",
			ConnectionTimeout: "1s",
			RetryTimeout:      "2s",
			RetryInterval:     "3s",
			PrivateKey:        "doh",
		},
	}

	e := []interface{}{
		map[string]interface{}{
			"direct": []interface{}{
				map[string]interface{}{},
			},
			"ssh": []interface{}{
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
			},
		},
	}

	if diff := cmp.Diff(hostMarshal(c, true), e); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestHostUnmarshal(t *testing.T) {
	c := host.Host{
		DirectConfig: &direct.Config{},
		SSHConfig: &ssh.Config{
			Address:           "127.0.0.1",
			Port:              examplePort,
			User:              "foo",
			Password:          "bar",
			ConnectionTimeout: "1s",
			RetryTimeout:      "2s",
			RetryInterval:     "3s",
			PrivateKey:        "doh",
		},
	}

	e := []interface{}{
		map[string]interface{}{
			"direct": []interface{}{
				map[string]interface{}{},
			},
			"ssh": []interface{}{
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
			},
		},
	}

	if diff := cmp.Diff(hostUnmarshal(e[0]), c); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestHostUnmarshalEmpty(t *testing.T) {
	h := host.Host{
		DirectConfig: &direct.Config{},
	}

	if diff := cmp.Diff(hostUnmarshal(nil), h); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestHostUnmarshalEmptyBock(t *testing.T) {
	h := host.Host{
		DirectConfig: &direct.Config{},
	}

	if diff := cmp.Diff(hostUnmarshal(map[string]interface{}{}), h); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}
