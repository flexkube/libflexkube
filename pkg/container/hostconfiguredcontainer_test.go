package container

import (
	"fmt"
	"net"
	"testing"

	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

// withHook()
func TestWithHook(t *testing.T) {
	action := false

	if err := withHook(nil, func() error {
		action = true

		return nil
	}, nil); err != nil {
		t.Fatalf("withHook should not return error, got: %v", err)
	}

	if !action {
		t.Fatalf("withHook should execute action")
	}
}

func TestWithPreHook(t *testing.T) {
	pre := false

	f := Hook(func() error {
		pre = true

		return nil
	})

	if err := withHook(&f, func() error {
		return nil
	}, nil); err != nil {
		t.Fatalf("withHook should not return error, got: %v", err)
	}

	if !pre {
		t.Fatalf("withHook should call pre-hook")
	}
}

func TestWithPostHook(t *testing.T) {
	post := false

	f := Hook(func() error {
		post = true

		return nil
	})

	if err := withHook(nil, func() error {
		return nil
	}, &f); err != nil {
		t.Fatalf("withHook should not return error, got: %v", err)
	}

	if !post {
		t.Fatalf("withHook should call post-hook")
	}
}

func TestConnectAndForward(t *testing.T) {
	addr := &net.UnixAddr{
		Name: "@foo",
		Net:  "unix",
	}

	localSock, err := net.ListenUnix("unix", addr)
	if err != nil {
		t.Fatalf("unable to listen on address '%s':%v", addr, err)
	}
	defer localSock.Close()

	h := &hostConfiguredContainer{
		host: host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	s, err := h.connectAndForward(fmt.Sprintf("unix://%s", addr.String()))
	if err != nil {
		t.Fatalf("Direct forwarding to open listener should work, got: %v", err)
	}

	if s == "" {
		t.Fatalf("Returned forwarded address shouldn't be empty")
	}
}
