package ssh

import (
	"reflect"
	"testing"
)

func TestBuildConfig(t *testing.T) {
	cases := []struct {
		config   *Config
		defaults *Config
		result   *Config
	}{
		// All defaults
		{
			nil,
			nil,
			&Config{
				Port:              Port,
				User:              User,
				ConnectionTimeout: ConnectionTimeout,
			},
		},

		// PrivateKey
		{
			&Config{
				PrivateKey: "foo",
			},
			nil,
			&Config{
				PrivateKey:        "foo",
				Port:              Port,
				User:              User,
				ConnectionTimeout: ConnectionTimeout,
			},
		},
		{
			&Config{
				PrivateKey: "foo",
			},
			&Config{
				PrivateKey: "bar",
			},
			&Config{
				PrivateKey:        "foo",
				Port:              Port,
				User:              User,
				ConnectionTimeout: ConnectionTimeout,
			},
		},
		{
			nil,
			&Config{
				PrivateKey: "bar",
			},
			&Config{
				PrivateKey:        "bar",
				Port:              Port,
				User:              User,
				ConnectionTimeout: ConnectionTimeout,
			},
		},

		// User
		{
			&Config{
				User: "foo",
			},
			nil,
			&Config{
				User:              "foo",
				Port:              Port,
				ConnectionTimeout: ConnectionTimeout,
			},
		},
		{
			&Config{
				User: "foo",
			},
			&Config{
				User: "bar",
			},
			&Config{
				User:              "foo",
				Port:              Port,
				ConnectionTimeout: ConnectionTimeout,
			},
		},
		{
			nil,
			&Config{
				User: "bar",
			},
			&Config{
				User:              "bar",
				Port:              Port,
				ConnectionTimeout: ConnectionTimeout,
			},
		},

		// ConnectionTimeout
		{
			&Config{
				ConnectionTimeout: "foo",
			},
			nil,
			&Config{
				ConnectionTimeout: "foo",
				Port:              Port,
				User:              User,
			},
		},
		{
			&Config{
				ConnectionTimeout: "foo",
			},
			&Config{
				ConnectionTimeout: "bar",
			},
			&Config{
				ConnectionTimeout: "foo",
				Port:              Port,
				User:              User,
			},
		},
		{
			nil,
			&Config{
				ConnectionTimeout: "bar",
			},
			&Config{
				ConnectionTimeout: "bar",
				Port:              Port,
				User:              User,
			},
		},

		// Port
		{
			&Config{
				Port: 33,
			},
			nil,
			&Config{
				ConnectionTimeout: ConnectionTimeout,
				Port:              33,
				User:              User,
			},
		},
		{
			&Config{
				Port: 33,
			},
			&Config{
				Port: 44,
			},
			&Config{
				ConnectionTimeout: ConnectionTimeout,
				Port:              33,
				User:              User,
			},
		},
		{
			nil,
			&Config{
				Port: 33,
			},
			&Config{
				ConnectionTimeout: ConnectionTimeout,
				Port:              33,
				User:              User,
			},
		},
	}

	for _, c := range cases {
		if nc := BuildConfig(c.config, c.defaults); !reflect.DeepEqual(nc, c.result) {
			t.Fatalf("expected %+v, got %+v", c.result, nc)
		}
	}
}
