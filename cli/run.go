package cli

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/types"
)

const (
	ExitError = 1
	ExitOK    = 0
)

func Run(rc types.ResourceConfig) int {
	// Read files.
	f, err := readFiles()
	if err != nil {
		fmt.Printf("Failed reading files: %v\n", err)

		return ExitError
	}

	c, err := prepare(f, rc)
	if err != nil {
		fmt.Printf("Failed preparing object: %v\n", err)

		return ExitError
	}

	// Calculate and print diff.
	fmt.Printf("Calculating diff...\n\n")

	d := cmp.Diff(c.Containers().ToExported().PreviousState, c.Containers().DesiredState())

	if d == "" {
		fmt.Println("No changes required")

		return ExitOK
	}

	fmt.Printf("Following changes required:\n\n%s\n\n", d)

	return deploy(c)
}

// readYamlFile reads YAML file from disk and handles empty files,
// so they can be merged.
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

// Read files reads state and config files from disk and returns their
// content merged.
func readFiles() ([]byte, error) {
	s, err := readYamlFile("state.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed reading state file: %w", err)
	}

	fmt.Println("Reading config file config.yaml")

	config, err := readYamlFile("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed reading config file: %w", err)
	}

	return []byte(string(s) + string(config)), nil
}

func saveState(c types.Resource) error {
	fmt.Println("Saving state to state.yaml file")

	state, err := c.StateToYaml()
	if err != nil {
		return fmt.Errorf("failed serializing new state: %w", err)
	}

	if string(state) == "{}\n" {
		state = []byte{}
	}

	if err := ioutil.WriteFile("state.yaml", state, 0644); err != nil {
		return fmt.Errorf("failed writing new state to file: %w", err)
	}

	return nil
}

func prepare(f []byte, rc types.ResourceConfig) (types.Resource, error) {
	// Create object.
	fmt.Println("Creating object")

	c, err := types.ResourceFromYaml(f, rc)
	if err != nil {
		return nil, fmt.Errorf("failed parsing config file: %w", err)
	}

	// Check current state.
	fmt.Println("Checking current state")

	if err := c.CheckCurrentState(); err != nil {
		return nil, fmt.Errorf("failed checking current state: %w", err)
	}

	return c, nil
}

func deploy(c types.Resource) int {
	fmt.Println("Deploying updates")

	deployErr := c.Deploy()

	if err := saveState(c); err != nil {
		fmt.Printf("Failed persisting state: %v\n", err)

		return ExitError
	}

	if deployErr != nil {
		fmt.Printf("Deploying failed: %v\n", deployErr)

		return ExitError
	}

	fmt.Println("Done")

	return ExitOK
}
