package flexkube

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/container/types"
)

func TestPortsMarshal(t *testing.T) {
	c := []types.PortMap{
		{
			IP:       "127.0.0.1",
			Port:     examplePort,
			Protocol: "tcp",
		},
	}

	e := []interface{}{
		map[string]interface{}{
			"ip":       "127.0.0.1",
			"port":     examplePort,
			"protocol": "tcp",
		},
	}

	if diff := cmp.Diff(portMapMarshal(c), e); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestPortsUnmarshal(t *testing.T) {
	c := []types.PortMap{
		{
			IP:       "127.0.0.1",
			Port:     examplePort,
			Protocol: "tcp",
		},
	}

	e := []interface{}{
		map[string]interface{}{
			"ip":       "127.0.0.1",
			"port":     examplePort,
			"protocol": "tcp",
		},
	}

	if diff := cmp.Diff(portMapUnmarshal(e), c); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}
