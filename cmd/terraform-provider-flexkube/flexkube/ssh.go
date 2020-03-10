package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
)

func sshMarshal(c ssh.Config, sensitive bool) interface{} {
	p := c.Password
	if sensitive && c.Password != "" {
		p = sha256sum([]byte(c.Password))
	}

	pk := c.PrivateKey
	if sensitive && c.PrivateKey != "" {
		pk = sha256sum([]byte(c.PrivateKey))
	}

	return []interface{}{
		map[string]interface{}{
			"address":            c.Address,
			"port":               c.Port,
			"user":               c.User,
			"password":           p,
			"connection_timeout": c.ConnectionTimeout,
			"retry_timeout":      c.RetryTimeout,
			"retry_interval":     c.RetryInterval,
			"private_key":        pk,
		},
	}
}

func sshUnmarshal(i interface{}) *ssh.Config {
	// If block is not defined, don't return anything.
	if i == nil {
		return nil
	}

	j, ok := i.(map[string]interface{})
	if !ok || len(j) == 0 {
		return nil
	}

	c := &ssh.Config{}

	if h, ok := j["address"]; ok {
		c.Address = h.(string)
	}

	if h, ok := j["port"]; ok {
		c.Port = h.(int)
	}

	if h, ok := j["user"]; ok {
		c.User = h.(string)
	}

	if h, ok := j["password"]; ok {
		c.Password = h.(string)
	}

	if h, ok := j["connection_timeout"]; ok {
		c.ConnectionTimeout = h.(string)
	}

	if h, ok := j["retry_timeout"]; ok {
		c.RetryTimeout = h.(string)
	}

	if h, ok := j["retry_interval"]; ok {
		c.RetryInterval = h.(string)
	}

	if h, ok := j["private_key"]; ok {
		c.PrivateKey = h.(string)
	}

	return c
}

func sshSchema(computed bool) *schema.Schema {
	return optionalBlock(computed, func(computed bool) map[string]*schema.Schema {
		return map[string]*schema.Schema{
			"address":            optionalString(computed),
			"port":               optionalInt(computed),
			"user":               optionalString(computed),
			"password":           sensitiveString(computed), // sensitive!
			"connection_timeout": optionalString(computed),
			"retry_timeout":      optionalString(computed),
			"retry_interval":     optionalString(computed),
			"private_key":        sensitiveString(computed), // sensitive!
		}
	})
}
