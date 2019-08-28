package main

import (
	"fmt"
	"io/ioutil"

	"github.com/invidian/etcd-ariadnes-thread/pkg"
	"github.com/invidian/etcd-ariadnes-thread/pkg/defaults"
	"github.com/invidian/etcd-ariadnes-thread/pkg/node"
)

func main() {
	etcd := etcd.New()
	if err := etcd.SetName("clusterA"); err != nil {
		panic(err)
	}
	if err := etcd.SetImage(defaults.EtcdImage); err != nil {
		panic(err)
	}
	node := &node.Node{
		Name: "foo",
	}
	if err := etcd.AddNode(node); err != nil {
		panic(err)
	}

	previousState, err := ioutil.ReadFile("./state.json")
	if err := etcd.LoadPreviousState(previousState); err != nil {
		fmt.Println("Previous state not found.")
	}

	if err := etcd.ReadCurrentState(); err != nil {
		panic(err)
	}
	if err := etcd.Plan(); err != nil {
		panic(err)
	}
	etcd.PresentPlan()
	if err := etcd.Apply(); err != nil {
		panic(err)
	}

	currentState, err := etcd.DumpCurrentState()
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile("./state.json", currentState, 0644); err != nil {
		panic(err)
	}
}
