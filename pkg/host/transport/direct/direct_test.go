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
