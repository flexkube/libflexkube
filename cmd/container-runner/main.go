package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/invidian/libflexkube/pkg/container"
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

func saveStateFile(data []byte) error {
	return ioutil.WriteFile("state.yaml", data, 0644)
}

func main() {
	s, err := readYamlFile("state.yaml")
	if err != nil {
		panic(err)
	}
	config, err := readYamlFile("config.yaml")
	if err != nil {
		panic(err)
	}
	c, err := container.FromYaml([]byte(string(s) + string(config)))
	if err != nil {
		panic(err)
	}
	fmt.Println("Checking current state...\n")
	if err := c.CheckCurrentState(); err != nil {
		panic(err)
	}
	fmt.Println("Applying changes...\n")
	if err := c.Execute(); err != nil {
		panic(err)
	}
	fmt.Println("")
	fmt.Println("Dumping current state to file...\n")
	state, err := c.CurrentStateToYaml()
	if err != nil {
		panic(err)
	}
	if string(state) == "{}\n" {
		state = []byte{}
	}
	if err := ioutil.WriteFile("state.yaml", state, 0644); err != nil {
		panic(err)
	}
	fmt.Println("Done")
}
