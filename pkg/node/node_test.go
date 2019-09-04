package node

import (
	"testing"

	"github.com/invidian/etcd-ariadnes-thread/pkg/container"
)

// New()
func TestNewEmptyConfiguration(t *testing.T) {
	if _, err := New(&Node{}); err == nil {
		t.Errorf("Creating node with wrong configuration should fail")
	}
}

func TestNewGoodConfiguration(t *testing.T) {
	n := &Node{
		Name:          "foo",
		ContainerName: "bar",
	}
	if _, err := New(n); err != nil {
		t.Errorf("Creating node with good configuration should pass, got: %v", err)
	}
}

func TestNewValidateContainerRuntime(t *testing.T) {
	n := &Node{
		Name:                 "foo",
		ContainerName:        "bar",
		ContainerRuntimeName: "doh",
	}
	if _, err := New(n); err == nil {
		t.Errorf("Creating node should validate container runtime configuration")
	}
}

// Validate()
func TestValidateNoName(t *testing.T) {
	n := &Node{
		ContainerName: "foo",
	}
	if err := n.Validate(); err == nil {
		t.Errorf("Validating node without name should fail")
	}
}

func TestValidateNoContainerName(t *testing.T) {
	n := &Node{
		Name: "foo",
	}
	if err := n.Validate(); err == nil {
		t.Errorf("Validating node without container name should fail")
	}
}

func TestValidate(t *testing.T) {
	n := &Node{
		Name:          "foo",
		ContainerName: "foo",
	}
	if err := n.Validate(); err != nil {
		t.Errorf("Validating node with valid configuration should pass, got: %v", err)
	}
}

func TestValidateUnsupportedRuntime(t *testing.T) {
	n := &Node{
		Name:                 "foo",
		ContainerName:        "foo",
		ContainerRuntimeName: "foo",
	}
	if err := n.Validate(); err == nil {
		t.Errorf("Validating node with unsupported container runtime should fail")
	}
}

// selectContainerRuntime()
func TestSelectDockerContainerRuntime(t *testing.T) {
	n := &node{
		nodeBase{
			containerRuntimeName: "docker",
		},
		&container.Status{},
	}
	if err := n.selectContainerRuntime(); err != nil {
		t.Errorf("Selecting Docker container runtime should succeed, got %v", err)
	}
	if n.containerRuntime == nil {
		t.Errorf("Selecting container runtime should set container runtime field")
	}
}

func TestSelectDefaultContainerRuntime(t *testing.T) {
	n := &node{}
	if err := n.selectContainerRuntime(); err != nil {
		t.Errorf("Selecting container runtime on empty node should succeed, got %v", err)
	}
	if n.containerRuntime == nil {
		t.Errorf("Selecting container runtime should set container runtime field")
	}
}

func TestSelectBadContainerRuntime(t *testing.T) {
	n := &node{
		nodeBase{
			containerRuntimeName: "foo",
		},
		&container.Status{},
	}
	if err := n.selectContainerRuntime(); err == nil {
		t.Errorf("Unsupported container runtime name should be rejected")
	}
}
