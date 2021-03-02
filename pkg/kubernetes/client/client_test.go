package client

import (
	"errors"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

func TestCheckNodeExistsFakeKubeconfig(t *testing.T) {
	t.Parallel()

	kubeconfig := GetKubeconfig(t)

	c, err := NewClient([]byte(kubeconfig))
	if err != nil {
		t.Fatalf("Failed creating client: %v", err)
	}

	e, err := c.CheckNodeExists("foo")()

	if e == true {
		t.Errorf("Node should never exists with fake kubeconfig")
	}

	if err == nil {
		t.Errorf("Checking node existence should always fail with fake kubeconfig")
	}
}

func TestWaitForNodeFakeKubeconfig(t *testing.T) {
	t.Parallel()

	kubeconfig := GetKubeconfig(t)

	c, err := NewClient([]byte(kubeconfig))
	if err != nil {
		t.Fatalf("Failed creating client: %v", err)
	}

	if err := c.WaitForNode("foo"); err == nil {
		t.Errorf("Waiting for node should always fail with fake kubeconfig")
	}
}

func TestLabelNodeFakeKubeconfig(t *testing.T) {
	t.Parallel()

	kubeconfig := GetKubeconfig(t)

	c, err := NewClient([]byte(kubeconfig))
	if err != nil {
		t.Fatalf("Failed creating client: %v", err)
	}

	l := map[string]string{
		"foo": "bar",
	}

	if err := c.LabelNode("foo", l); err == nil {
		t.Errorf("Labeling node should always fail with fake kubeconfig")
	}
}

// PingWait() tests.
func TestPingWaitFakeKubeconfig(t *testing.T) {
	t.Parallel()

	kubeconfig := GetKubeconfig(t)

	c, err := NewClient([]byte(kubeconfig))
	if err != nil {
		t.Fatalf("Failed creating client: %v", err)
	}

	if err := c.PingWait(1*time.Second, 1*time.Second); !errors.Is(err, wait.ErrWaitTimeout) {
		t.Fatalf("Ping with fake config should always timeout, got: %v", err)
	}
}

// CheckNodeReady() tests.
func TestCheckNodeReadyFakeKubeconfig(t *testing.T) {
	t.Parallel()

	kubeconfig := GetKubeconfig(t)

	c, err := NewClient([]byte(kubeconfig))
	if err != nil {
		t.Fatalf("Failed creating client: %v", err)
	}

	e, err := c.CheckNodeReady("foo")()

	if e == true {
		t.Errorf("check should never return true with fake kubeconfig")
	}

	if err != nil {
		t.Errorf("check should swallow all errors and just return boolean value")
	}
}
