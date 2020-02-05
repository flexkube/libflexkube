package direct

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	d := &Config{}

	di, err := d.New()
	if err != nil {
		t.Fatalf("should return new object without errors, got: %v", err)
	}

	if !reflect.DeepEqual(di, &direct{}) {
		t.Fatalf("should be equal to empty struct, got: %+v", di)
	}
}

func TestForwardUnixSocket(t *testing.T) {
	d := &direct{}
	p := "/foo"

	if fp, _ := d.ForwardUnixSocket(p); fp != p {
		t.Fatalf("expected '%s', got '%s'", p, fp)
	}
}

func TestConnect(t *testing.T) {
	d := &direct{}
	if _, err := d.Connect(); err != nil {
		t.Fatalf("Connect should always work, got: %v", err)
	}
}

func TestForwardTCP(t *testing.T) {
	d := &direct{}
	a := "localhost:80"

	if fa, _ := d.ForwardTCP(a); fa != a {
		t.Fatalf("expected '%s', got '%s'", a, fa)
	}
}

func TestForwardTCPBadAddress(t *testing.T) {
	d := &direct{}
	a := "localhost"

	if _, err := d.ForwardTCP(a); err == nil {
		t.Fatalf("TCP forwarding should fail when forwarding bad address")
	}
}
