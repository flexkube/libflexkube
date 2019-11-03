package helm

import (
	"fmt"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"sigs.k8s.io/yaml"
)

// Release represents user-configured Helm release
type Release struct {

	// Kubeconfig is content of kubeconfig file in YAML format, which will be used to authenticate
	// to the cluster and create a release.
	Kubeconfig string `json:"kubeconfig,omitempty" yaml:"kubeconfig,omitempty"`

	// Namespace is a namespace, where helm release will be created and all it's resources
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`

	// Name is a name of the release used to indentify it
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Chart is a location of the chart. It may be local path or remote chart in user repository
	Chart string `json:"chart,omitempty" yaml:"chart,omitempty"`

	// Values is a chart values in YAML format
	Values string `json:"values,omitempty" yaml:"values,omitempty"`

	// Version is a requested version of the chart
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

// release is a validated and installable/update'able version of Release
type release struct {
	actionConfig *action.Configuration
	settings     *cli.EnvSettings
	values       map[string]interface{}
	name         string
	namespace    string
	version      string
	chart        string
}

func (r *Release) New() (*release, error) {
	if err := r.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate helm release: %w", err)
	}

	// Initialize kubernetes and helm CLI clients
	actionConfig := &action.Configuration{}
	settings := cli.New()

	// Safe to ignore errors, becuase Validate will return early if data is not valid
	g, kc, cs, _ := newClients(r.Kubeconfig)

	actionConfig.RESTClientGetter = g
	actionConfig.KubeClient = kc
	actionConfig.Releases = storage.Init(driver.NewSecrets(cs.CoreV1().Secrets(r.Namespace)))

	values, _ := r.parseValues()

	release := &release{
		actionConfig: actionConfig,
		settings:     settings,
		values:       values,
		name:         r.Name,
		namespace:    r.Namespace,
		version:      r.Version,
		chart:        r.Chart,
	}

	return release, nil
}

func (r *Release) Validate() error {
	// Check if all required values are filled in
	if r.Kubeconfig == "" {
		return fmt.Errorf("kubeconfig is empty")
	}
	if r.Namespace == "" {
		return fmt.Errorf("namespace is empty")
	}
	if r.Name == "" {
		return fmt.Errorf("name is empty")
	}
	if r.Chart == "" {
		return fmt.Errorf("chart is empty")
	}

	// Try to create a clients
	if _, _, _, err := newClients(r.Kubeconfig); err != nil {
		return fmt.Errorf("failed to create kubernetes clients: %w", err)
	}

	// Parse given values
	if _, err := r.parseValues(); err != nil {
		return fmt.Errorf("failed to parse values: %w", err)
	}

	return nil
}

// ValidateChart locates and parses the chart.
//
// This method is not part of Validate(), since Validate() should be fully offline and no-op.
// However, if user wants know that the chart is already available and wants to avoid runtime
// errors, this function can be called in addition to Validate().
func (r *release) ValidateChart() error {
	client := r.installClient()

	if _, err := r.loadChart(client); err != nil {
		return fmt.Errorf("failed validating chart: %w", err)
	}

	return nil
}

// Install installs configured chart as release
func (r *release) Install() error {
	client := r.installClient()

	chart, err := r.loadChart(client)
	if err != nil {
		return fmt.Errorf("loading chart failed: %w", err)
	}

	// Install a release
	if _, err = client.Run(chart, r.values); err != nil {
		return fmt.Errorf("installing a release failed: %w", err)
	}

	return nil
}

// loadChart locates and loads the chart
func (r *release) loadChart(client *action.Install) (*chart.Chart, error) {
	// Locate chart to install, here we install chart from local folder
	cp, err := client.ChartPathOptions.LocateChart(r.chart, r.settings)
	if err != nil {
		return nil, fmt.Errorf("locating chart failed: %w", err)
	}

	return loader.Load(cp)
}

// installClient returns action install client for helm
func (r *release) installClient() *action.Install {
	// Initialize install action client
	// TODO maybe there is more generic action we could use?
	client := action.NewInstall(r.actionConfig)

	client.Version = r.version
	client.ReleaseName = r.name
	client.Namespace = r.namespace

	return client
}

// parseValues parses release values and returns it ready to use when installing chart
func (r *Release) parseValues() (map[string]interface{}, error) {
	values := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(r.Values), &values); err != nil {
		return nil, fmt.Errorf("failed to parse values: %w", err)
	}

	return values, nil
}
