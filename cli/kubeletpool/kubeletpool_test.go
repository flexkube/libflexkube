package kubeletpool

import (
	"testing"

	"github.com/flexkube/libflexkube/cli"
)

func TestRun(t *testing.T) {
	if i := Run(); i != cli.ExitError {
		t.Fatalf("Running CLI app with no configuration should fail")
	}
}
