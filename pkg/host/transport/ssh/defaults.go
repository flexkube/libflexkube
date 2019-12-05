package ssh

import (
	"github.com/flexkube/libflexkube/internal/util"
)

const (
	// User is a default user used for SSH connections.
	User = "root"

	// ConnectionTimeout is a default time SSH will wait while connecting to unreachable server.
	ConnectionTimeout = "30s"

	// RetryTimeout is a default time after we give up connecting to reachable server.
	RetryTimeout = "60s"

	// RetryInterval is a default time how long we wait between SSH connection attempts.
	RetryInterval = "1s"

	// Port is a default port used for SSH connections.
	Port = 22
)

// BuildConfig takes destination SSH configuration, struct with default values provided by the user
// and merges it together with global SSH default values.
func BuildConfig(sshConfig *Config, defaults *Config) *Config {
	if sshConfig == nil {
		sshConfig = &Config{}
	}

	if defaults == nil {
		defaults = &Config{}
	}

	sshConfig.PrivateKey = util.PickString(sshConfig.PrivateKey, defaults.PrivateKey)

	sshConfig.User = util.PickString(sshConfig.User, defaults.User, User)

	sshConfig.ConnectionTimeout = util.PickString(sshConfig.ConnectionTimeout, defaults.ConnectionTimeout, ConnectionTimeout)

	sshConfig.RetryTimeout = util.PickString(sshConfig.RetryTimeout, defaults.RetryTimeout, RetryTimeout)

	sshConfig.RetryInterval = util.PickString(sshConfig.RetryInterval, defaults.RetryInterval, RetryInterval)

	sshConfig.Port = util.PickInt(sshConfig.Port, defaults.Port, Port)

	return sshConfig
}
