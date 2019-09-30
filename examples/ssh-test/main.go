package main

import (
	"fmt"

	"github.com/invidian/etcd-ariadnes-thread/pkg/host/transport/ssh"
)

func main() {
	config := ssh.SSHConfig{
		Host:              "localhost",
		Port:              22,
		User:              "invidian",
		Password:          "foobarbaz",
		ConnectionTimeout: "30s",
	}
	t, err := config.New()
	if err != nil {
		panic(err)
	}

	s, err := t.ForwardUnixSocket("unix:///run/docker.sock")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Forwarded to %s\n", s)
}
