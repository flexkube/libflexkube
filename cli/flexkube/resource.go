package flexkube

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"text/template"

	sprig "github.com/Masterminds/sprig/v3"
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
	// Etcd allows to manage etcd cluster, which is required for running Kubernetes.
	//
	// See etcd.Cluster for available fields.
	Etcd *etcd.Cluster `json:"etcd,omitempty"`

	// Controlplane allows to manage static Kubernetes controlplane, which consists of kube-apiserver,
	// kube-scheduler and kube-controller-manager.
	//
	// Usually single controlplane is created, which can be then used to install self-hosted controlplane
	// e.g. using 'helm'.
	//
	// See controlplane.Controlplane for available fields.
	Controlplane *controlplane.Controlplane `json:"controlplane,omitempty"`

	// PKI allows to manage certificates and private keys required by Kubernetes.
	//
	// See pki.PKI for available fields.
	PKI *pki.PKI `json:"pki,omitempty"`

	// KubeletPools allows to manage multiple kubelet pools. In case of self-hosted Kubernetes, you usually want to
	// run at least 2 pools, which differs in labels. One for controller nodes, which runs for example kube-apiserver
	// and other pods, which have access to cluster credentials and another one for worker nodes.
	//
	// Creating more worker pools is useful, if you have group of cluster nodes with different hardware.
	//
	// See kubelet.Pool for available fields.
	KubeletPools map[string]*kubelet.Pool `json:"kubeletPools,omitempty"`

	// APILoadBalancerPools allows to manage multiple kube-apiserver load balancers, which allows building
	// highly available Kubernetes clusters.
	//
	// For example, kubelet does not support load-balancing internally, so it can be pointed to load balancer
	// address, which will handle graceful handover in case of one API server going down.
	//
	// See apiloadbalancer.APILoadBalancers for available fields.
	APILoadBalancerPools map[string]*apiloadbalancer.APILoadBalancers `json:"apiLoadBalancerPools,omitempty"`

	// Containers allows to manage arbitrary container groups. This is useful, when you need some extra
	// containers to run as part of your Kubernetes cluster. For example, when running with cloud-controller-manager
	// it can be used to run static instance of it for bootstrapping.
	//
	// See container.ContainersState for available options.
	Containers map[string]*container.ContainersState `json:"containers,omitempty"`

	// State stores state of all configured resources. Information about all created containers and generated certificates
	// must be persisted, so it does not change on consecutive runs.
	State *ResourceState `json:"state,omitempty"`

	// Confirmed controls, if user should be asked for confirmation input before applying changes.
	// Set to 'true' for unattended runs.
	Confirmed bool `json:"confirmed,omitempty"`

	// Noop controls, if deployment should actually be executed. If set to 'true', only the difference between
	// cluster existing state and desired state will be printed, but the State field won't be modified.
	Noop bool `json:"noop,omitempty"`
}

// ResourceState represents flexkube CLI state format.
type ResourceState struct {
	// Etcd stores state information about containers which are part of etcd cluster.
	Etcd *container.ContainersState `json:"etcd,omitempty"`

	// Controlplane stores state information about containers which are part Kubernetes static controlplane.
	Controlplane *container.ContainersState `json:"controlplane,omitempty"`

	// KubeletPools stores state information about containers which are part of kubelet pools.
	KubeletPools map[string]*container.ContainersState `json:"kubeletPools,omitempty"`

	// APILoadBalancerPools stores state information about containers which are part of kube-apiserver load
	// balancer pools.
	APILoadBalancerPools map[string]*container.ContainersState `json:"apiLoadBalancerPools,omitempty"`

	// Containers stores state information for configured container groups.
	Containers map[string]*container.ContainersState `json:"containers,omitempty"`

	// PKI stores generated Kubernetes certificates.
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
		return nil, fmt.Errorf("serializing PKI configuration: %w", err)
	}

	if err := yaml.Unmarshal(pkic, pki); err != nil {
		return nil, fmt.Errorf("merging PKI configuration with state: %w", err)
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

	if config == nil {
		config = &container.ContainersState{}
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
		return nil, fmt.Errorf("validating configuration: %w", err)
	}

	r, err := rc.New()
	if err != nil {
		return nil, fmt.Errorf("initializing object: %w", err)
	}

	return r, nil
}

func checkState(resource types.Resource) (string, error) {
	// Check current state.
	fmt.Println("Checking current state")

	if err := resource.CheckCurrentState(); err != nil {
		return "", fmt.Errorf("checking current state: %w", err)
	}

	// Calculate and print diff.
	fmt.Printf("Calculating diff...\n\n")

	diff := cmp.Diff(resource.Containers().ToExported().PreviousState, resource.Containers().DesiredState())

	if diff == "" {
		fmt.Println("No changes required")

		return diff, nil
	}

	fmt.Printf("Following changes required:\n\n%s\n\n", util.ColorizeDiff(diff))

	return diff, nil
}

// execute checks current state of the deployment and triggers the deployment if needed.
func (r *Resource) execute(resource types.Resource, saveStateF func(types.Resource)) error {
	diff, err := checkState(resource)
	if err != nil {
		return fmt.Errorf("checking current state: %w", err)
	}

	if r.Noop || diff == "" {
		return nil
	}

	return r.deploy(resource, saveStateF)
}

// deploy confirms the deployment with the user and persists the state after the deployment.
func (r *Resource) deploy(resource types.Resource, saveStateF func(types.Resource)) error {
	if !r.Confirmed {
		confirmed, err := askForConfirmation()
		if err != nil {
			return fmt.Errorf("asking for confirmation: %w", err)
		}

		if !confirmed {
			fmt.Println("Aborted")

			return nil
		}
	}

	deployErr := resource.Deploy()

	if r.State == nil {
		r.State = &ResourceState{}
	}

	saveStateF(resource)

	return r.StateToFile(deployErr)
}

func askForConfirmation() (bool, error) {
	r := bufio.NewReader(os.Stdin)

	fmt.Printf("To continue, type (y)es nad press enter: ")

	response, err := r.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("reading user response: %w", err)
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
	configRaw, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	// Workaround for empty YAML file.
	if string(configRaw) == "{}\n" {
		return []byte{}, nil
	}

	return configRaw, nil
}

// LoadResourceFromFiles loads Resource struct from config.yaml and state.yaml files.
func LoadResourceFromFiles() (*Resource, error) {
	resource := &Resource{}

	configRaw, err := readYamlFile("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("reading config.yaml file: %w", err)
	}

	stateRaw, err := readYamlFile("state.yaml")
	if err != nil {
		return nil, fmt.Errorf("reading state.yaml file: %w", err)
	}

	if err := yaml.Unmarshal([]byte(string(configRaw)+string(stateRaw)), resource); err != nil {
		return nil, fmt.Errorf("parsing files: %w", err)
	}

	return resource, nil
}

// StateToFile saves resource state into state.yaml file.
func (r *Resource) StateToFile(actionErr error) error {
	rs := &Resource{
		State: r.State,
	}

	stateRaw, err := yaml.Marshal(rs)
	if err != nil {
		return fmt.Errorf("serializing state: %w", err)
	}

	if string(stateRaw) == "{}\n" {
		stateRaw = []byte{}
	}

	readWriteOwnerOnly := 0o600

	// #nosec G115 // Constant conversion.
	if err := os.WriteFile("state.yaml", stateRaw, fs.FileMode(readWriteOwnerOnly)); err != nil {
		if actionErr == nil {
			return fmt.Errorf("writing new state to file: %w", err)
		}

		fmt.Println("Failed to write state.yaml file: %w", err)
	}

	if actionErr != nil {
		return fmt.Errorf("executing action: %w", actionErr)
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
		//nolint:stylecheck // Kubernetes is a proper noun so should be capitalized.
		return fmt.Errorf("Kubernetes PKI management not enabled")
	}

	if r.State.PKI.Kubernetes.AdminCertificate == nil {
		//nolint:stylecheck // Kubernetes is a proper noun so should be capitalized.
		return fmt.Errorf("Kubernetes admin certificate not available in PKI")
	}

	return nil
}

// validateKubeconfigControlplane validates if required fields are populated in Controlplane
// configuration to generate admin kubeconfig file.
func (r *Resource) validateKubeconfigControlplane() error {
	if r.Controlplane == nil {
		//nolint:stylecheck // Kubernetes is a proper noun so should be capitalized.
		return fmt.Errorf("Kubernetes controlplane management not enabled")
	}

	if r.Controlplane.APIServerAddress == "" {
		//nolint:stylecheck // Kubernetes is a proper noun so should be capitalized.
		return fmt.Errorf("Kubernetes controlplane has no API server address set")
	}

	if r.Controlplane.APIServerPort == 0 {
		//nolint:stylecheck // Kubernetes is a proper noun so should be capitalized.
		return fmt.Errorf("Kubernetes controlplane has no API server port set")
	}

	return nil
}

// validateKubeconfig validates, if kubeconfig content can be generated from current
// state of the resource.
func (r *Resource) validateKubeconfig() error {
	if err := r.validateKubeconfigPKI(); err != nil {
		return fmt.Errorf("validating PKI fields required for generating kubeconfig: %w", err)
	}

	if err := r.validateKubeconfigControlplane(); err != nil {
		return fmt.Errorf("validating controlplane fields required for generating kubeconfig: %w", err)
	}

	return nil
}

// Kubeconfig generates content of kubeconfig file in YAML format from Controlplane and PKI
// configuration.
func (r *Resource) Kubeconfig() (string, error) {
	if err := r.validateKubeconfig(); err != nil {
		return "", fmt.Errorf("validating kubeconfig: %w", err)
	}

	clientConfig := &client.Config{
		Server:            fmt.Sprintf("%s:%d", r.Controlplane.APIServerAddress, r.Controlplane.APIServerPort),
		CACertificate:     r.State.PKI.Kubernetes.CA.X509Certificate,
		ClientCertificate: r.State.PKI.Kubernetes.AdminCertificate.X509Certificate,
		ClientKey:         r.State.PKI.Kubernetes.AdminCertificate.PrivateKey,
	}

	k, err := clientConfig.ToYAMLString()
	if err != nil {
		return "", fmt.Errorf("generating client configuration: %w", err)
	}

	return k, nil
}

// RunAPILoadBalancerPool deploys given API Load Balancer pool.
func (r *Resource) RunAPILoadBalancerPool(name string) error {
	pool, err := r.getAPILoadBalancerPool(name)
	if err != nil {
		return fmt.Errorf("getting API Load Balancer pool %q from configuration: %w", name, err)
	}

	saveStateF := func(types.Resource) {
		if r.State.APILoadBalancerPools == nil {
			r.State.APILoadBalancerPools = map[string]*container.ContainersState{}
		}

		r.State.APILoadBalancerPools[name] = &pool.Containers().ToExported().PreviousState
	}

	return r.execute(pool, saveStateF)
}

// RunControlplane deploys configured static controlplane.
func (r *Resource) RunControlplane() error {
	controlplaneResource, err := r.getControlplane()
	if err != nil {
		return fmt.Errorf("getting controlplane from the configuration: %w", err)
	}

	saveStateF := func(types.Resource) {
		r.State.Controlplane = &controlplaneResource.Containers().ToExported().PreviousState
	}

	return r.execute(controlplaneResource, saveStateF)
}

// RunEtcd deploys configured etcd cluster.
func (r *Resource) RunEtcd() error {
	etcdResource, err := r.getEtcd()
	if err != nil {
		return fmt.Errorf("getting etcd from the configuration: %w", err)
	}

	saveStateF := func(types.Resource) {
		r.State.Etcd = &etcdResource.Containers().ToExported().PreviousState
	}

	return r.execute(etcdResource, saveStateF)
}

// RunKubeletPool deploys given kubelet pool.
func (r *Resource) RunKubeletPool(name string) error {
	kubeletPool, err := r.getKubeletPool(name)
	if err != nil {
		return fmt.Errorf("getting kubelet pool %q from configuration: %w", name, err)
	}

	saveStateF := func(types.Resource) {
		if r.State.KubeletPools == nil {
			r.State.KubeletPools = map[string]*container.ContainersState{}
		}

		r.State.KubeletPools[name] = &kubeletPool.Containers().ToExported().PreviousState
	}

	return r.execute(kubeletPool, saveStateF)
}

// RunPKI generates configured PKI.
func (r *Resource) RunPKI() error {
	pki, err := r.getPKI()
	if err != nil {
		return fmt.Errorf("loading PKI configuration: %w", err)
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
	containersResource, err := r.getContainers(name)
	if err != nil {
		return fmt.Errorf("getting containers group %q from configuration: %w", name, err)
	}

	saveStateF := func(types.Resource) {
		if r.State.Containers == nil {
			r.State.Containers = map[string]*container.ContainersState{}
		}

		r.State.Containers[name] = &containersResource.Containers().ToExported().PreviousState
	}

	return r.execute(containersResource, saveStateF)
}

// Template executes given Go template using configuration and state.
func (r *Resource) Template(templateContent string) (string, error) {
	tmpl, err := template.New("template").Funcs(sprig.TxtFuncMap()).Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, r); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}

// TemplateFromFile reads template from a given path and executes it using configuration and state.
func (r *Resource) TemplateFromFile(templatePath string) (string, error) {
	t, err := os.ReadFile(templatePath) // #nosec G304
	if err != nil {
		return "", fmt.Errorf("reading template file %q: %w", templatePath, err)
	}

	return r.Template(string(t))
}
