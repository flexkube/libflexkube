// Package flexkube contains logic of 'flexkube' CLI.
package flexkube

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/urfave/cli/v2"
)

const (
	// Version is a version printed by the --version flag.
	Version = "v0.6.1-unreleased"

	// YesFlag is a const for --yes flag.
	YesFlag = "yes"

	// NoopFlag is const for --noop flag.
	NoopFlag = "noop"
)

// Run executes flexkube CLI binary with given arguments (usually os.Args).
func Run(args []string) int {
	app := &cli.App{
		Name:    "flexkube",
		Version: Version,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  YesFlag,
				Usage: "Evaluate the configuration without confirmation",
			},
			&cli.BoolFlag{
				Name:  NoopFlag,
				Usage: "Only checks the status of the deployment, but does not do any changes",
			},
		},
		Commands: []*cli.Command{
			kubeletPoolCommand(),
			apiLoadBalancerPoolCommand(),
			etcdCommand(),
			pkiCommand(),
			controlplaneCommand(),
			kubeconfigCommand(),
			containersCommand(),
			templateCommand(),
		},
	}

	if err := app.Run(args); err != nil {
		fmt.Printf("Execution failed: %v\n", err)

		return 1
	}

	return 0
}

func templateCommand() *cli.Command {
	return &cli.Command{
		Name:      "template",
		Usage:     "reads Go template from given file or stdin and evaluates it using configuration and state",
		ArgsUsage: "[TEMPLATE FILE PATH]",
		Action: func(c *cli.Context) error {
			return withResource(c, templateAction)
		},
	}
}

func kubeletPoolCommand() *cli.Command {
	return &cli.Command{
		Name:      "kubelet-pool",
		Usage:     "executes kubelet pool configuration",
		ArgsUsage: "[POOL NAME]",
		Action: func(c *cli.Context) error {
			return withResource(c, kubeletPoolAction)
		},
	}
}

func apiLoadBalancerPoolCommand() *cli.Command {
	return &cli.Command{
		Name:      "apiloadbalancer-pool",
		Usage:     "executes API Load Balancer pool configuration",
		ArgsUsage: "[POOL NAME]",
		Action: func(c *cli.Context) error {
			return withResource(c, apiLoadBalancerPoolAction)
		},
	}
}

func etcdCommand() *cli.Command {
	return &cli.Command{
		Name:  "etcd",
		Usage: "execute etcd configuration",
		Action: func(c *cli.Context) error {
			return withResource(c, etcdAction)
		},
	}
}

func pkiCommand() *cli.Command {
	return &cli.Command{
		Name:  "pki",
		Usage: "execute PKI configuration",
		Action: func(c *cli.Context) error {
			return withResource(c, pkiAction)
		},
	}
}

func controlplaneCommand() *cli.Command {
	return &cli.Command{
		Name:  "controlplane",
		Usage: "execute controlplane configuration",
		Action: func(c *cli.Context) error {
			return withResource(c, controlplaneAction)
		},
	}
}

func kubeconfigCommand() *cli.Command {
	return &cli.Command{
		Name:  "kubeconfig",
		Usage: "prints admin kubeconfig for cluster",
		Action: func(c *cli.Context) error {
			return withResource(c, kubeconfigAction)
		},
	}
}

func containersCommand() *cli.Command {
	return &cli.Command{
		Name:  "containers",
		Usage: "manages arbitrary container pools",
		Action: func(c *cli.Context) error {
			return withResource(c, containersAction)
		},
	}
}

// apiLoadBalancerPoolAction implements 'apiloadbalancer-pool' subcommand.
func apiLoadBalancerPoolAction(c *cli.Context, r *Resource) error {
	poolName, err := getPoolName(c)
	if err != nil {
		return fmt.Errorf("getting pool name: %w", err)
	}

	return r.RunAPILoadBalancerPool(poolName)
}

// controlplaneAction implements 'controlplane' subcommand.
func controlplaneAction(c *cli.Context, r *Resource) error {
	return r.RunControlplane()
}

// etcdAction implements 'etcd' subcommand.
func etcdAction(c *cli.Context, r *Resource) error {
	return r.RunEtcd()
}

// getTemplate reads the template either from path given as an argument
// or from stdin.
func getTemplate(c *cli.Context) (string, error) {
	if c.NArg() > 1 {
		return "", fmt.Errorf("only one template file can be evaluated at a time")
	}

	template := []byte{}

	if c.NArg() == 1 {
		p := c.Args().Get(0)

		c, err := ioutil.ReadFile(p) // #nosec G304
		if err != nil {
			return "", fmt.Errorf("reading template file %q: %w", p, err)
		}

		template = c
	}

	if c.NArg() == 0 {
		bytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading template from stdin: %w", err)
		}

		template = bytes
	}

	return string(template), nil
}

// templateActions runs Resource.Template().
func templateAction(c *cli.Context, r *Resource) error {
	template, err := getTemplate(c)
	if err != nil {
		return fmt.Errorf("getting template: %w", err)
	}

	o, err := r.Template(template)
	if err != nil {
		return fmt.Errorf("templating: %w", err)
	}

	fmt.Println(o)

	return nil
}

func kubeconfigAction(c *cli.Context, r *Resource) error {
	k, err := r.Kubeconfig()
	if err != nil {
		return fmt.Errorf("failed generating kubeconfig: %w", err)
	}

	fmt.Println(k)

	return nil
}

func kubeletPoolAction(c *cli.Context, r *Resource) error {
	poolName, err := getPoolName(c)
	if err != nil {
		return fmt.Errorf("getting pool name %w", err)
	}

	return r.RunKubeletPool(poolName)
}

func pkiAction(c *cli.Context, r *Resource) error {
	return r.RunPKI()
}

func getPoolName(c *cli.Context) (string, error) {
	if c.NArg() > 1 {
		return "", fmt.Errorf("only one pool can be managed at a time")
	}

	poolName := c.Args().Get(0)
	if poolName == "" {
		return "", fmt.Errorf("pool name must be specified")
	}

	return poolName, nil
}

func containersAction(c *cli.Context, r *Resource) error {
	poolName, err := getPoolName(c)
	if err != nil {
		return fmt.Errorf("getting pool name: %w", err)
	}

	return r.RunContainers(poolName)
}

// withResource is a helper for action functions.
func withResource(c *cli.Context, rf func(*cli.Context, *Resource) error) error {
	r, err := LoadResourceFromFiles()
	if err != nil {
		return fmt.Errorf("reading configuration and state failed: %w", err)
	}

	r.Confirmed = c.Bool(YesFlag)
	r.Noop = c.Bool(NoopFlag)

	if r.Confirmed && r.Noop {
		return fmt.Errorf("--%s and --%s flags are mutually exclusive", YesFlag, NoopFlag)
	}

	if r.Noop {
		fmt.Println("No-op run, no changes will be made.")
	}

	return rf(c, r)
}
