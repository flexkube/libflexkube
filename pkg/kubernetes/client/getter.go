package client

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// Getter implements k8s.io/cli-runtime/pkg/genericclioptions.RESTClientGetter interface.
type Getter struct {
	c clientcmd.ClientConfig
}

// ToRESTMapper is part of k8s.io/cli-runtime/pkg/genericclioptions.RESTClientGetter interface.
func (c *Getter) ToRESTMapper() (meta.RESTMapper, error) {
	d, err := c.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(d)
	expander := restmapper.NewShortcutExpander(mapper, d)

	return expander, nil
}

// ToDiscoveryClient is part of k8s.io/cli-runtime/pkg/genericclioptions.RESTClientGetter interface.
func (c *Getter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
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

// ToRawKubeConfigLoader is part of k8s.io/cli-runtime/pkg/genericclioptions.RESTClientGetter interface.
func (c *Getter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return c.c
}

// ToRESTConfig is part of k8s.io/cli-runtime/pkg/genericclioptions.RESTClientGetter interface.
func (c *Getter) ToRESTConfig() (*rest.Config, error) {
	return c.c.ClientConfig()
}

// NewGetter takes content of kubeconfig file as an argument and returns implementation of
// RESTClientGetter k8s interface.
func NewGetter(data []byte) (*Getter, error) {
	c, err := clientcmd.NewClientConfigFromBytes(data)
	if err != nil {
		return nil, err
	}

	return &Getter{
		c: c,
	}, nil
}
