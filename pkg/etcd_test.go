package etcd

import (
	"reflect"
	"testing"

	"github.com/invidian/etcd-ariadnes-thread/pkg/node"
	"github.com/invidian/etcd-ariadnes-thread/pkg/step"
)

func TestNewCluster(t *testing.T) {
	etcd := New()
	if etcd.PreviousState == nil {
		t.Error("Previous state is empty in new cluster")
	}

	if etcd.currentState != nil {
		t.Error("current state should be empty in new cluster")
	}

	if etcd.DesiredState == nil {
		t.Error("Desired state is empty in new cluster")
	}
}

func TestAddNewClusterNode(t *testing.T) {
	etcd := New()
	node := &node.Node{
		Name: "foo",
	}
	if err := etcd.AddNode(node); err != nil {
		t.Errorf("Adding new node should not fail, got: %s", err)
	}
	if etcd.DesiredState.Nodes[node.Name] == nil {
		t.Errorf("Adding new node should not fail")
	}
}

func TestReadEmptyClusterState(t *testing.T) {
	etcd := New()
	if err := etcd.ReadCurrentState(); err != nil {
		t.Errorf("Reading state of empty cluster should succeed, got: %s", err)
	}
	if etcd.currentState == nil {
		t.Errorf("Reading current state should set current state")
	}
}

func TestReadClusterState(t *testing.T) {
	etcd := New()
	node := &node.Node{
		Name: "foo",
	}
	if err := etcd.AddNode(node); err != nil {
		t.Errorf("Adding new node should not fail, got: %s", err)
	}
	if err := etcd.ReadCurrentState(); err != nil {
		t.Errorf("Reading state of empty cluster should succeed, got: %s", err)
	}
	if etcd.currentState == nil {
		t.Errorf("Reading cluster state should set current state")
	}
	if etcd.currentState.Nodes[node.Name] != nil {
		t.Errorf("Current state should not have node set on fresh cluster")
	}
}

func TestPlanOnUnknownCluster(t *testing.T) {
	etcd := New()
	if err := etcd.Plan(); err == nil {
		t.Errorf("Planning should fail on cluster without read state")
	}
}

func TestPlanCluster(t *testing.T) {
	etcd := New()
	if err := etcd.ReadCurrentState(); err != nil {
		t.Errorf("Reading state of empty cluster should succeed, got: %s", err)
	}
	if err := etcd.Plan(); err != nil {
		t.Errorf("Planning should succeed, got: %s", err)
	}
}

func TestSetImage(t *testing.T) {
	etcd := New()
	image := "gcr.io/etcd-development/etcd:v3.3.13"
	if err := etcd.SetImage(image); err != nil {
		t.Errorf("Setting image should succeed, got: %s", err)
	}
	if etcd.DesiredState.Image != image {
		t.Errorf("Desired Image should be '%s', got: '%s'", image, etcd.DesiredState.Image)
	}
}

func TestPlanClusterWithOneNode(t *testing.T) {
	etcd := New()
	node := &node.Node{
		Name: "foo",
	}
	if err := etcd.AddNode(node); err != nil {
		t.Errorf("Adding new node should not fail, got: %s", err)
	}
	if err := etcd.ReadCurrentState(); err != nil {
		t.Errorf("Reading state of empty cluster should succeed, got: %s", err)
	}
	if err := etcd.Plan(); err != nil {
		t.Errorf("Planning should succeed, got: %s", err)
	}
	steps := step.Steps{
		&step.Step{
			StepType: step.AddNode,
			Node:     node,
		},
	}
	if !reflect.DeepEqual(etcd.Steps, steps) {
		t.Errorf("Plan should contain one AddNode step")
	}
}

func TestPlanClusterRemoveOneNode(t *testing.T) {
	etcd := New()
	node := &node.Node{
		Name: "foo",
	}
	etcd.PreviousState.Nodes[node.Name] = node
	if err := etcd.ReadCurrentState(); err != nil {
		t.Errorf("Reading state of empty cluster should succeed, got: %s", err)
	}
	if err := etcd.Plan(); err != nil {
		t.Errorf("Planning should succeed, got: %s", err)
	}
	steps := step.Steps{
		&step.Step{
			StepType: step.RemoveNode,
			Node:     node,
		},
	}
	if !reflect.DeepEqual(etcd.Steps, steps) {
		t.Errorf("Plan should contain one RemoveNode step")
	}
}
