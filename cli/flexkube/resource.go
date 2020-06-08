package flexkube

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/pkg/apiloadbalancer"
	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/resource"
	"github.com/flexkube/libflexkube/pkg/controlplane"
	"github.com/flexkube/libflexkube/pkg/etcd"
	"github.com/flexkube/libflexkube/pkg/kubelet"
	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
	"github.com/flexkube/libflexkube/pkg/pki"
	"github.com/flexkube/libflexkube/pkg/types"
)

// Resource represents flexkube CLI configuration structure.
type Resource struct {
	Etcd                 *etcd.Cluster                                `json:"etcd,omitempty"`
	Controlplane         *controlplane.Controlplane                   `json:"controlplane,omitempty"`
	PKI                  *pki.PKI                                     `json:"pki,omitempty"`
	KubeletPools         map[string]*kubelet.Pool                     `json:"kubeletPools,omitempty"`
	APILoadBalancerPools map[string]*apiloadbalancer.APILoadBalancers `json:"apiLoadBalancerPools,omitempty"`
	State                *ResourceState                               `json:"state,omitempty"`
	Confirmed            bool                                         `json:"confirmed,omitempty"`
	Noop                 bool                                         `json:"noop,omitempty"`
}

// ResourceState represents flexkube CLI state format.
type ResourceState struct {
	Etcd                 *container.ContainersState            `json:"etcd,omitempty"`
	Controlplane         *container.ContainersState            `json:"controlplane,omitempty"`
	KubeletPools         map[string]*container.ContainersState `json:"kubeletPools,omitempty"`
	APILoadBalancerPools map[string]*container.ContainersState `json:"apiLoadBalancerPools,omitempty"`

	// Containers stores state information for configured container groups.
	Containers map[string]*container.ContainersState `json:"containers,omitempty"`

	PKI *pki.PKI `json:"pki,omitempty"`
}

// getEtcd returns etcd resource, with state and PKI integration enabled.
func (r *Resource) getEtcd() (types.Resource, error) {
	if r.Etcd == nil {
		if r.State == nil || r.State.Etcd == nil {
			return nil, fmt.Errorf("etcd management not enabled in the configuration and state not found")
		}

		r.Etcd = &etcd.Cluster{}
	}

	if r.State != nil && r.State.Etcd != nil {
		r.Etcd.State = *r.State.Etcd
	}

	// Enable PKI integration.
	if r.State != nil && r.State.PKI != nil {
		r.Etcd.PKI = r.State.PKI
	}

	return validateAndNew(r.Etcd)
}

// getControlplane returns controlplane resource, with state and PKI integration enabled.
func (r *Resource) getControlplane() (types.Resource, error) {
	if r.Controlplane == nil {
		if r.State == nil || r.State.Controlplane == nil {
			return nil, fmt.Errorf("controlplane not configured and state not found")
		}

		r.Controlplane = &controlplane.Controlplane{
			Destroy: true,
		}
	}

	if r.State != nil {
		r.Controlplane.State = r.State.Controlplane
	}

	// Enable PKI integration.
	if r.State != nil && r.State.PKI != nil {
		r.Controlplane.PKI = r.State.PKI
	}

	return validateAndNew(r.Controlplane)
}

// getKubeletPool returns requested kubelet pool with state and PKI injected.
func (r *Resource) getKubeletPool(name string) (types.Resource, error) {
	stateFound := r.State != nil && r.State.KubeletPools != nil && r.State.KubeletPools[name] != nil
	configPool, configFound := r.KubeletPools[name]

	if !stateFound && !configFound {
		return nil, fmt.Errorf("pool not configured and state not found")
	}

	pool := &kubelet.Pool{}

	if configFound {
		pool = configPool
	}

	if stateFound {
		pool.State = *r.State.KubeletPools[name]
	}

	// Enable PKI integration.
	if r.State != nil && r.State.PKI != nil {
		pool.PKI = r.State.PKI
	}

	return validateAndNew(pool)
}

// getPKI returns PKI struct with state loaded on top.
func (r *Resource) getPKI() (*pki.PKI, error) {
	if r.PKI == nil {
		return nil, fmt.Errorf("PKI config configured")
	}

	pki := &pki.PKI{}

	// If state contains PKI, use it as a base for loading.
	if r.State != nil && r.State.PKI != nil {
		fmt.Println("Loading existing PKI state from state.yaml file")

		pki = r.State.PKI
	}

	// Then load config on top.
	pkic, err := yaml.Marshal(r.PKI)
	if err != nil {
		return nil, fmt.Errorf("serializing PKI configuration failed: %w", err)
	}

	if err := yaml.Unmarshal(pkic, pki); err != nil {
		return nil, fmt.Errorf("failed merging PKI configuration with state: %w", err)
	}

	return pki, nil
}

// getAPILoadBalancerPool returns requested kubelet pool with state injected.
func (r *Resource) getAPILoadBalancerPool(name string) (types.Resource, error) {
	stateFound := r.State != nil && r.State.APILoadBalancerPools != nil && r.State.APILoadBalancerPools[name] != nil
	configPool, configFound := r.APILoadBalancerPools[name]

	if !stateFound && !configFound {
		return nil, fmt.Errorf("pool not configured and state not found")
	}

	pool := &apiloadbalancer.APILoadBalancers{}

	if configFound {
		pool = configPool
	}

	if stateFound {
		pool.State = *r.State.APILoadBalancerPools[name]
	}

	return validateAndNew(pool)
}

// getContainers returns requested containers group with state.
func (r *Resource) getContainers(name string) (types.Resource, error) {
	stateFound := r.State != nil && r.State.Containers != nil && r.State.Containers[name] != nil
	config, configFound := r.Containers[name]

	if !stateFound && !configFound {
		return nil, fmt.Errorf("group not configured and state not found")
	}

	containers := &resource.Containers{}

	if configFound {
		containers.Containers = *config
	}

	if stateFound {
		containers.State = *r.State.Containers[name]
	}

	return validateAndNew(containers)
}

// validateAndNew validates and creates new resource from resource config.
func validateAndNew(rc types.ResourceConfig) (types.Resource, error) {
	if err := rc.Validate(); err != nil {
		return nil, fmt.Errorf("validating configuration failed: %w", err)
	}

	r, err := rc.New()
	if err != nil {
		return nil, fmt.Errorf("initializing object failed: %w", err)
	}

	return r, nil
}

func (r *Resource) checkState(rs types.Resource) (string, error) {
	// Check current state.
	fmt.Println("Checking current state")

	if err := rs.CheckCurrentState(); err != nil {
		return "", fmt.Errorf("failed checking current state: %w", err)
	}

	// Calculate and print diff.
	fmt.Printf("Calculating diff...\n\n")

	d := cmp.Diff(rs.Containers().ToExported().PreviousState, rs.Containers().DesiredState())

	if d == "" {
		fmt.Println("No changes required")

		return d, nil
	}

	fmt.Printf("Following changes required:\n\n%s\n\n", util.ColorizeDiff(d))

	return d, nil
}

// execute checks current state of the deployment and triggers the deployment if needed.
func (r *Resource) execute(rs types.Resource, saveStateF func(types.Resource)) error {
	diff, err := r.checkState(rs)
	if err != nil {
		return fmt.Errorf("failed checking current state: %w", err)
	}

	if r.Noop || diff == "" {
		return nil
	}

	return r.deploy(rs, saveStateF)
}

// deploy confirms the deployment with the user and persists the state after the deployment.
func (r *Resource) deploy(rs types.Resource, saveStateF func(types.Resource)) error {
	if !r.Confirmed {
		confirmed, err := askForConfirmation()
		if err != nil {
			return fmt.Errorf("failed asking for confirmation: %w", err)
		}

		if !confirmed {
			fmt.Println("Aborted")

			return nil
		}
	}

	deployErr := rs.Deploy()

	if r.State == nil {
		r.State = &ResourceState{}
	}

	saveStateF(rs)

	return r.StateToFile(deployErr)
}

func askForConfirmation() (bool, error) {
	r := bufio.NewReader(os.Stdin)

	fmt.Printf("To continue, type (y)es nad press enter: ")

	response, err := r.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed reading user response: %w", err)
	}

	switch strings.ToLower(strings.TrimSpace(response)) {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return askForConfirmation()
	}
}

// readYamlFile reads YAML file from disk and handles empty files,
// so they can be merged.
func readYamlFile(file string) ([]byte, error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return []byte(""), nil
	}

	// The function is not exported and all parameters to this function
	// are static.
	//
	// #nosec G304
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

// LoadResourceFromFiles loads Resource struct from config.yaml and state.yaml files.
func LoadResourceFromFiles() (*Resource, error) {
	r := &Resource{}

	c, err := readYamlFile("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("reading config.yaml file failed: %w", err)
	}

	s, err := readYamlFile("state.yaml")
	if err != nil {
		return nil, fmt.Errorf("reading state.yaml file failed: %w", err)
	}

	if err := yaml.Unmarshal([]byte(string(c)+string(s)), r); err != nil {
		return nil, fmt.Errorf("parsing files failed: %w", err)
	}

	return r, nil
}

// StateToFile saves resource state into state.yaml file.
func (r *Resource) StateToFile(actionErr error) error {
	rs := &Resource{
		State: r.State,
	}

	rb, err := yaml.Marshal(rs)
	if err != nil {
		return fmt.Errorf("failed serializing state: %w", err)
	}

	if string(rb) == "{}\n" {
		rb = []byte{}
	}

	if err := ioutil.WriteFile("state.yaml", rb, 0600); err != nil {
		if actionErr == nil {
			return fmt.Errorf("failed writing new state to file: %w", err)
		}

		fmt.Println("Failed to write state.yaml file: %w", err)
	}

	if actionErr != nil {
		return fmt.Errorf("execution failed: %w", actionErr)
	}

	fmt.Println("Action complete")

	return nil
}

// validateKubeconfigPKI validates if required fields are populated in PKI field
// to generate admin kubeconfig file.
func (r *Resource) validateKubeconfigPKI() error {
	if r.State.PKI == nil {
		return fmt.Errorf("PKI management not enabled")
	}

	if r.State.PKI.Kubernetes == nil {
		return fmt.Errorf("Kubernetes PKI management not enabled") //nolint:stylecheck
	}

	if r.State.PKI.Kubernetes.AdminCertificate == nil {
		return fmt.Errorf("Kubernetes admin certificate not available in PKI") //nolint:stylecheck
	}

	return nil
}

// validateKubeconfigControlplane validates if required fields are populated in PKI field
// to generate admin kubeconfig file.
func (r *Resource) validateKubeconfigControlplane() error {
	if r.Controlplane == nil {
		return fmt.Errorf("Kubernetes controlplane management not enabled") //nolint:stylecheck
	}

	if r.Controlplane.APIServerAddress == "" {
		return fmt.Errorf("Kubernetes controlplane has no API server address set") //nolint:stylecheck
	}

	if r.Controlplane.APIServerPort == 0 {
		return fmt.Errorf("Kubernetes controlplane has no API server port set") //nolint:stylecheck
	}

	return nil
}

// validateKubeconfig validates, if kubeconfig content can be generated from current
// state of the resource.
func (r *Resource) validateKubeconfig() error {
	if err := r.validateKubeconfigPKI(); err != nil {
		return err
	}

	if err := r.validateKubeconfigControlplane(); err != nil {
		return err
	}

	return nil
}

// Kubeconfig generates content of kubeconfig file in YAML format from Controlplane and PKI
// configuration.
func (r *Resource) Kubeconfig() (string, error) {
	if err := r.validateKubeconfig(); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	cc := &client.Config{
		Server:            fmt.Sprintf("%s:%d", r.Controlplane.APIServerAddress, r.Controlplane.APIServerPort),
		CACertificate:     r.State.PKI.Kubernetes.CA.X509Certificate,
		ClientCertificate: r.State.PKI.Kubernetes.AdminCertificate.X509Certificate,
		ClientKey:         r.State.PKI.Kubernetes.AdminCertificate.PrivateKey,
	}

	k, err := cc.ToYAMLString()
	if err != nil {
		return "", fmt.Errorf("generating failed: %w", err)
	}

	return k, nil
}

// RunAPILoadBalancerPool deploys given API Load Balancer pool.
func (r *Resource) RunAPILoadBalancerPool(name string) error {
	p, err := r.getAPILoadBalancerPool(name)
	if err != nil {
		return fmt.Errorf("failed getting API Load Balancer pool %q from configuration: %w", name, err)
	}

	saveStateF := func(rs types.Resource) {
		if r.State.APILoadBalancerPools == nil {
			r.State.APILoadBalancerPools = map[string]*container.ContainersState{}
		}

		r.State.APILoadBalancerPools[name] = &p.Containers().ToExported().PreviousState
	}

	return r.execute(p, saveStateF)
}

// RunControlplane deploys configured static controlplane.
func (r *Resource) RunControlplane() error {
	e, err := r.getControlplane()
	if err != nil {
		return fmt.Errorf("failed getting controlplane from the configuration: %w", err)
	}

	saveStateF := func(rs types.Resource) {
		r.State.Controlplane = &e.Containers().ToExported().PreviousState
	}

	return r.execute(e, saveStateF)
}

// RunEtcd deploys configured etcd cluster.
func (r *Resource) RunEtcd() error {
	e, err := r.getEtcd()
	if err != nil {
		return fmt.Errorf("preparing failed: %w", err)
	}

	saveStateF := func(rs types.Resource) {
		r.State.Etcd = &e.Containers().ToExported().PreviousState
	}

	return r.execute(e, saveStateF)
}

// RunKubeletPool deploys given kubelet pool.
func (r *Resource) RunKubeletPool(name string) error {
	p, err := r.getKubeletPool(name)
	if err != nil {
		return fmt.Errorf("failed getting kubelet pool %q from configuration: %w", name, err)
	}

	saveStateF := func(rs types.Resource) {
		if r.State.KubeletPools == nil {
			r.State.KubeletPools = map[string]*container.ContainersState{}
		}

		r.State.KubeletPools[name] = &p.Containers().ToExported().PreviousState
	}

	return r.execute(p, saveStateF)
}

// RunPKI generates configured PKI.
func (r *Resource) RunPKI() error {
	pki, err := r.getPKI()
	if err != nil {
		return fmt.Errorf("failed loading PKI configuration: %w", err)
	}

	fmt.Println("Generating PKI...")

	genErr := pki.Generate()

	if r.State == nil {
		r.State = &ResourceState{}
	}

	r.State.PKI = pki

	return r.StateToFile(genErr)
}

// RunContainers deploys given containers group.
func (r *Resource) RunContainers(name string) error {
	p, err := r.getContainers(name)
	if err != nil {
		return fmt.Errorf("failed getting containers group %q from configuration: %w", name, err)
	}

	saveStateF := func(rs types.Resource) {
		if r.State.Containers == nil {
			r.State.Containers = map[string]*container.ContainersState{}
		}

		r.State.Containers[name] = &p.Containers().ToExported().PreviousState
	}

	return r.execute(p, saveStateF)
}
