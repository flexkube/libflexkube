package ssh_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
)

const (
	// Custom port for testing, which differs from default SSH port.
	customPort = 33
)

func TestBuildConfig(t *testing.T) { //nolint:funlen // Just many test cases.
	t.Parallel()

	cases := []struct {
		config   *ssh.Config
		defaults *ssh.Config
		result   *ssh.Config
	}{
		// All defaults
		{
			nil,
			nil,
			&ssh.Config{
				Port:              ssh.Port,
				User:              ssh.User,
				ConnectionTimeout: ssh.ConnectionTimeout,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     ssh.RetryInterval,
			},
		},

		// PrivateKey
		{
			&ssh.Config{
				PrivateKey: "foo",
			},
			nil,
			&ssh.Config{
				PrivateKey:        "foo",
				Port:              ssh.Port,
				User:              ssh.User,
				ConnectionTimeout: ssh.ConnectionTimeout,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     ssh.RetryInterval,
			},
		},
		{
			&ssh.Config{
				PrivateKey: "foo",
			},
			&ssh.Config{
				PrivateKey: "bar",
			},
			&ssh.Config{
				PrivateKey:        "foo",
				Port:              ssh.Port,
				User:              ssh.User,
				ConnectionTimeout: ssh.ConnectionTimeout,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     ssh.RetryInterval,
			},
		},
		{
			nil,
			&ssh.Config{
				PrivateKey: "bar",
			},
			&ssh.Config{
				PrivateKey:        "bar",
				Port:              ssh.Port,
				User:              ssh.User,
				ConnectionTimeout: ssh.ConnectionTimeout,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     ssh.RetryInterval,
			},
		},

		// User
		{
			&ssh.Config{
				User: "foo",
			},
			nil,
			&ssh.Config{
				User:              "foo",
				Port:              ssh.Port,
				ConnectionTimeout: ssh.ConnectionTimeout,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     ssh.RetryInterval,
			},
		},
		{
			&ssh.Config{
				User: "foo",
			},
			&ssh.Config{
				User: "bar",
			},
			&ssh.Config{
				User:              "foo",
				Port:              ssh.Port,
				ConnectionTimeout: ssh.ConnectionTimeout,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     ssh.RetryInterval,
			},
		},
		{
			nil,
			&ssh.Config{
				User: "bar",
			},
			&ssh.Config{
				User:              "bar",
				Port:              ssh.Port,
				ConnectionTimeout: ssh.ConnectionTimeout,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     ssh.RetryInterval,
			},
		},

		// ConnectionTimeout
		{
			&ssh.Config{
				ConnectionTimeout: "foo",
			},
			nil,
			&ssh.Config{
				ConnectionTimeout: "foo",
				Port:              ssh.Port,
				User:              ssh.User,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     ssh.RetryInterval,
			},
		},
		{
			&ssh.Config{
				ConnectionTimeout: "foo",
			},
			&ssh.Config{
				ConnectionTimeout: "bar",
			},
			&ssh.Config{
				ConnectionTimeout: "foo",
				Port:              ssh.Port,
				User:              ssh.User,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     ssh.RetryInterval,
			},
		},
		{
			nil,
			&ssh.Config{
				ConnectionTimeout: "bar",
			},
			&ssh.Config{
				ConnectionTimeout: "bar",
				Port:              ssh.Port,
				User:              ssh.User,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     ssh.RetryInterval,
			},
		},

		// Port
		{
			&ssh.Config{
				Port: customPort,
			},
			nil,
			&ssh.Config{
				ConnectionTimeout: ssh.ConnectionTimeout,
				Port:              customPort,
				User:              ssh.User,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     ssh.RetryInterval,
			},
		},
		{
			&ssh.Config{
				Port: customPort,
			},
			&ssh.Config{
				Port: 44,
			},
			&ssh.Config{
				ConnectionTimeout: ssh.ConnectionTimeout,
				Port:              customPort,
				User:              ssh.User,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     ssh.RetryInterval,
			},
		},
		{
			nil,
			&ssh.Config{
				Port: customPort,
			},
			&ssh.Config{
				ConnectionTimeout: ssh.ConnectionTimeout,
				Port:              customPort,
				User:              ssh.User,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     ssh.RetryInterval,
			},
		},

		// RetryTimeout
		{
			&ssh.Config{
				RetryTimeout: "20s",
			},
			nil,
			&ssh.Config{
				ConnectionTimeout: ssh.ConnectionTimeout,
				Port:              ssh.Port,
				User:              ssh.User,
				RetryTimeout:      "20s",
				RetryInterval:     ssh.RetryInterval,
			},
		},
		{
			&ssh.Config{
				RetryTimeout: "20s",
			},
			&ssh.Config{
				RetryTimeout: "40s",
			},
			&ssh.Config{
				ConnectionTimeout: ssh.ConnectionTimeout,
				Port:              ssh.Port,
				User:              ssh.User,
				RetryTimeout:      "20s",
				RetryInterval:     ssh.RetryInterval,
			},
		},
		{
			nil,
			&ssh.Config{
				RetryTimeout: "40s",
			},
			&ssh.Config{
				ConnectionTimeout: ssh.ConnectionTimeout,
				Port:              ssh.Port,
				User:              ssh.User,
				RetryTimeout:      "40s",
				RetryInterval:     ssh.RetryInterval,
			},
		},

		// RetryInterval
		{
			&ssh.Config{
				RetryInterval: "5s",
			},
			nil,
			&ssh.Config{
				ConnectionTimeout: ssh.ConnectionTimeout,
				Port:              ssh.Port,
				User:              ssh.User,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     "5s",
			},
		},
		{
			&ssh.Config{
				RetryInterval: "5s",
			},
			&ssh.Config{
				RetryInterval: "10s",
			},
			&ssh.Config{
				ConnectionTimeout: ssh.ConnectionTimeout,
				Port:              ssh.Port,
				User:              ssh.User,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     "5s",
			},
		},
		{
			nil,
			&ssh.Config{
				RetryInterval: "5s",
			},
			&ssh.Config{
				ConnectionTimeout: ssh.ConnectionTimeout,
				Port:              ssh.Port,
				User:              ssh.User,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     "5s",
			},
		},

		// Address
		{
			&ssh.Config{
				Address: "localhost",
			},
			nil,
			&ssh.Config{
				ConnectionTimeout: ssh.ConnectionTimeout,
				Port:              ssh.Port,
				User:              ssh.User,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     ssh.RetryInterval,
				Address:           "localhost",
			},
		},
		{
			&ssh.Config{
				Address: "localhost",
			},
			&ssh.Config{
				Address: "foo",
			},
			&ssh.Config{
				ConnectionTimeout: ssh.ConnectionTimeout,
				Port:              ssh.Port,
				User:              ssh.User,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     ssh.RetryInterval,
				Address:           "localhost",
			},
		},
		{
			nil,
			&ssh.Config{
				Address: "localhost",
			},
			&ssh.Config{
				ConnectionTimeout: ssh.ConnectionTimeout,
				Port:              ssh.Port,
				User:              ssh.User,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     ssh.RetryInterval,
				Address:           "localhost",
			},
		},

		// Password
		{
			&ssh.Config{
				Password: "foo",
			},
			nil,
			&ssh.Config{
				ConnectionTimeout: ssh.ConnectionTimeout,
				Port:              ssh.Port,
				User:              ssh.User,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     ssh.RetryInterval,
				Password:          "foo",
			},
		},
		{
			&ssh.Config{
				Password: "foo",
			},
			&ssh.Config{
				Password: "bar",
			},
			&ssh.Config{
				ConnectionTimeout: ssh.ConnectionTimeout,
				Port:              ssh.Port,
				User:              ssh.User,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     ssh.RetryInterval,
				Password:          "foo",
			},
		},
		{
			nil,
			&ssh.Config{
				Password: "foo",
			},
			&ssh.Config{
				ConnectionTimeout: ssh.ConnectionTimeout,
				Port:              ssh.Port,
				User:              ssh.User,
				RetryTimeout:      ssh.RetryTimeout,
				RetryInterval:     ssh.RetryInterval,
				Password:          "foo",
			},
		},
	}

	for i, c := range cases {
		c := c

		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()

			if nc := ssh.BuildConfig(c.config, c.defaults); !reflect.DeepEqual(nc, c.result) {
				t.Fatalf("expected %+v, got %+v", c.result, nc)
			}
		})
	}
}
