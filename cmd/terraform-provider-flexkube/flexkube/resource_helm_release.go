package flexkube

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/yaml"
)

func resourceHelmRelease() *schema.Resource {
	return &schema.Resource{
		Create: resourceHelmReleaseCreate,
		Read:   resourceHelmReleaseRead,
		Delete: resourceHelmReleaseDelete,
		Update: resourceHelmReleaseCreate,
		Schema: map[string]*schema.Schema{
			"kubeconfig": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"namespace": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"chart": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"values": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceHelmReleaseCreate(d *schema.ResourceData, m interface{}) error {
	namespace := d.Get("namespace").(string)

	// Initialize kubernetes and helm CLI clients
	actionConfig := &action.Configuration{}
	settings := cli.New()

	// Inlining helm.sh/helm/v3/pkg/kube.New() to be able to override the config
	if err := v1beta1.AddToScheme(scheme.Scheme); err != nil {
		// This should never happen.
		panic(err)
	}

	getter, err := GetGetter([]byte(d.Get("kubeconfig").(string)))
	if err != nil {
		panic(err)
	}

	kc := &kube.Client{
		Factory: util.NewFactory(getter),
		Log:     func(_ string, _ ...interface{}) {},
	}

	clientset, err := kc.Factory.KubernetesClientSet()
	if err != nil {
		panic(err)
	}

	actionConfig.RESTClientGetter = getter
	actionConfig.KubeClient = kc
	actionConfig.Releases = storage.Init(driver.NewSecrets(clientset.CoreV1().Secrets(namespace)))

	// Initialize install action client
	client := action.NewInstall(actionConfig)

	// Set version of the chart
	client.Version = ">0.0.0-0"

	// Set release name
	client.ReleaseName = d.Get("name").(string)

	// Set namespace
	client.Namespace = namespace

	// Locate chart to install, here we install chart from local folder
	cp, err := client.ChartPathOptions.LocateChart(d.Get("chart").(string), settings)
	if err != nil {
		return fmt.Errorf("locating chart failed: %w", err)
	}

	values := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(d.Get("values").(string)), &values); err != nil {
		return fmt.Errorf("failed to parse values: %w", err)
	}

	// Load the chart
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return fmt.Errorf("loading chart failed: %w", err)
	}

	// Install a release
	if _, err = client.Run(chartRequested, values); err != nil {
		return fmt.Errorf("installing a release failed: %w", err)
	}

	result := sha256sum([]byte(d.Get("chart").(string) + d.Get("name").(string) + namespace + d.Get("kubeconfig").(string)))
	d.SetId(result)

	return nil
}

func resourceHelmReleaseRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceHelmReleaseDelete(d *schema.ResourceData, m interface{}) error {
	d.SetId("")
	return nil
}

// TODO All below should be moved to the library
type Client struct {
	c clientcmd.ClientConfig
}

// Implemented (part of k8s.io/cli-runtime/pkg/genericclioptions.RESTClientGetter)
func (c *Client) ToRESTMapper() (meta.RESTMapper, error) {
	d, err := c.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(d)
	expander := restmapper.NewShortcutExpander(mapper, d)
	return expander, nil
}

// Implemented (part of k8s.io/cli-runtime/pkg/genericclioptions.RESTClientGetter)
func (c *Client) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	cc, err := c.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	d, err := discovery.NewDiscoveryClientForConfig(cc)
	if err != nil {
		return nil, err
	}

	return memory.NewMemCacheClient(d), nil
}

// Implemented (part of k8s.io/cli-runtime/pkg/genericclioptions.RESTClientGetter)
func (c *Client) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return c.c
}

// Implemented (part of k8s.io/cli-runtime/pkg/genericclioptions.RESTClientGetter)
func (c *Client) ToRESTConfig() (*rest.Config, error) {
	return c.c.ClientConfig()
}

func GetGetter(data []byte) (*Client, error) {
	c, err := clientcmd.NewClientConfigFromBytes(data)
	if err != nil {
		return nil, err
	}

	return &Client{
		c: c,
	}, nil
}
