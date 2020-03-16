package client

import (
	"testing"
)

func TestCheckNodeExistsFakeKubeconfig(t *testing.T) {
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

// Ping()
func TestPingFakeKubeconfig(t *testing.T) {
	kubeconfig := GetKubeconfig(t)

	c, err := NewClient([]byte(kubeconfig))
	if err != nil {
		t.Fatalf("Failed creating client: %v", err)
	}

	e, err := c.(*client).Ping()

	if e == true {
		t.Errorf("Ping should never return true with fake kubeconfig")
	}

	if err != nil {
		t.Errorf("Ping should swallow all errors and just return boolean value")
	}
}
