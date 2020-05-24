package flexkube

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

const (
	// Version is a version printed by the --version flag.
	Version = "v0.3.0-unreleased"
)

// Run executes flexkube CLI binary with given arguments (usually os.Args).
func Run(args []string) int {
	app := &cli.App{
		Name:    "flexkube",
		Version: Version,
		Commands: []*cli.Command{
			{
				Name:      "kubelet-pool",
				Usage:     "executes kubelet pool configuration",
				ArgsUsage: "[POOL NAME]",
				Action:    withResource(kubeletPoolAction),
			},
			{
				Name:      "apiloadbalancer-pool",
				Usage:     "executes API Load Balancer pool configuration",
				ArgsUsage: "[POOL NAME]",
				Action:    withResource(apiLoadBalancerPoolAction),
			},
			{
				Name:   "etcd",
				Usage:  "execute etcd configuration",
				Action: withResource(etcdAction),
			},
			{
				Name:   "pki",
				Usage:  "execute PKI configuration",
				Action: withResource(pkiAction),
			},
			{
				Name:   "controlplane",
				Usage:  "execute controlplane configuration",
				Action: withResource(controlplaneAction),
			},
			{
				Name:   "kubeconfig",
				Usage:  "prints admin kubeconfig for cluster",
				Action: withResource(kubeconfigAction),
			},
		},
	}

	err := app.Run(args)
	if err != nil {
		fmt.Println(err.Error())

		return 1
	}

	return 0
}

// apiLoadBalancerPoolAction implements 'apiloadbalancer-pool' subcommand.
func apiLoadBalancerPoolAction(r *Resource) func(*cli.Context) error {
	return func(c *cli.Context) error {
		poolName, err := getPoolName(c)
		if err != nil {
			return err
		}

		return r.RunAPILoadBalancerPool(poolName)
	}
}

// controlplaneAction implements 'controlplane' subcommand.
func controlplaneAction(r *Resource) func(*cli.Context) error {
	return func(c *cli.Context) error {
		return r.RunControlplane()
	}
}

// etcdAction implements 'etcd' subcommand.
func etcdAction(r *Resource) func(*cli.Context) error {
	return func(c *cli.Context) error {
		return r.RunEtcd()
	}
}

func kubeconfigAction(r *Resource) func(*cli.Context) error {
	return func(c *cli.Context) error {
		k, err := r.Kubeconfig()
		if err != nil {
			return fmt.Errorf("failed generating kubeconfig: %w", err)
		}

		fmt.Println(k)

		return nil
	}
}

func kubeletPoolAction(r *Resource) func(*cli.Context) error {
	return func(c *cli.Context) error {
		poolName, err := getPoolName(c)
		if err != nil {
			return err
		}

		return r.RunKubeletPool(poolName)
	}
}

func pkiAction(r *Resource) func(*cli.Context) error {
	return func(c *cli.Context) error {
		return r.RunPKI()
	}
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
