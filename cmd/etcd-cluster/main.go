package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/flexkube/libflexkube/pkg/etcd"
)

func readYamlFile(file string) ([]byte, error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return []byte(""), nil
	}
	c, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	// Workaround for empty YAML file
	if string(c) == "{}\n" {
		return []byte{}, nil
	}
	return c, nil
}

func main() {
	fmt.Println("Reading state file state.yaml")
	s, err := readYamlFile("state.yaml")
	if err != nil {
		panic(err)
	}
	fmt.Println("Reading config file config.yaml")
	config, err := readYamlFile("config.yaml")
	if err != nil {
		panic(err)
	}
	fmt.Println("Creating etcd cluster object")
	c, err := etcd.FromYaml([]byte(string(s) + string(config)))
	if err != nil {
		panic(err)
	}
	fmt.Println("Checking current state of the cluster")
	if err := c.CheckCurrentState(); err != nil {
		panic(err)
	}
	fmt.Println("Deploying cluster updates")
	if err := c.Deploy(); err != nil {
		panic(err)
	}
	fmt.Println("Saving new cluster state to state.yaml file")
	state, err := c.StateToYaml()
	if err != nil {
		panic(err)
	}
	if string(state) == "{}\n" {
		state = []byte{}
	}
	if err := ioutil.WriteFile("state.yaml", state, 0644); err != nil {
		panic(err)
	}
	fmt.Println("Run complete")
}
