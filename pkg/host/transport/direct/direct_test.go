package direct_test

import (
	"testing"

	"github.com/flexkube/libflexkube/pkg/host/transport"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

func newDirect(t *testing.T) transport.Interface {
	t.Helper()

	d := &direct.Config{}

	di, err := d.New()
	if err != nil {
		t.Fatalf("Should return new object without errors, got: %v", err)
	}

	return di
}

func TestValidate(t *testing.T) {
	t.Parallel()

	d := &direct.Config{}

	if err := d.Validate(); err != nil {
		t.Fatalf("Validation should always pass, got: %v", err)
	}
}

func TestForwardUnixSocket(t *testing.T) {
	t.Parallel()

	d := newDirect(t)
	targetPath := "/foo"

	dc, err := d.Connect()
	if err != nil {
		t.Fatalf("Connecting: %v", err)
	}

	forwardedPath, err := dc.ForwardUnixSocket(targetPath)
	if err != nil {
		t.Fatalf("Forwarding socket: %v", err)
	}

	if forwardedPath != targetPath {
		t.Fatalf("Expected %q, got %q", targetPath, forwardedPath)
	}
}

func TestConnect(t *testing.T) {
	t.Parallel()

	d := newDirect(t)

	if _, err := d.Connect(); err != nil {
		t.Fatalf("Connect should always work, got: %v", err)
	}
}

func TestForwardTCP(t *testing.T) {
	t.Parallel()

	d := newDirect(t)
	targetAddress := "localhost:80"

	dc, err := d.Connect()
	if err != nil {
		t.Fatalf("Connecting: %v", err)
	}

	forwardedAddress, err := dc.ForwardTCP(targetAddress)
	if err != nil {
		t.Fatalf("Forwarding TCP: %v", err)
	}

	if forwardedAddress != targetAddress {
		t.Fatalf("Expected %q, got %q", targetAddress, forwardedAddress)
	}
}

func TestForwardTCPBadAddress(t *testing.T) {
	t.Parallel()

	d := newDirect(t)

	directConnected, err := d.Connect()
	if err != nil {
		t.Fatalf("Connecting: %v", err)
	}

	a := "localhost"

	if _, err := directConnected.ForwardTCP(a); err == nil {
		t.Fatalf("TCP forwarding should fail when forwarding bad address")
	}
}
