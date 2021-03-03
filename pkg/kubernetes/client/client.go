// Package client ships helper functions for building and using Kubernetes client.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

const (
	// PollInterval defines how long we wait before next attempt while waiting for the objects.
	PollInterval = 5 * time.Second

	// RetryTimeout defines how long we wait before timing out waiting for the objects.
	RetryTimeout = 10 * time.Minute
)

// Client defines exported capabilities of Flexkube k8s client.
type Client interface {
	// CheckNodeExists returns a function, which checks, if given node exists.
	CheckNodeExists(name string) func() (bool, error)

	// CheckNodeReady returns a function, which checks, if given node is ready.
	CheckNodeReady(name string) func() (bool, error)

	// WaitForNode waits, until Node object shows up in the API.
	WaitForNode(name string) error

	// WaitForNodeReady waits, until Node object becomes ready.
	WaitForNodeReady(name string) error

	// LabelNode patches Node object to set given labels on it.
	LabelNode(name string, labels map[string]string) error

	// PingWait waits until API server becomes available.
	PingWait(pollInterval, retryTimeout time.Duration) error
}

type client struct {
	*kubernetes.Clientset
}

// NewClient takes content of kubeconfig file as an argument and returns flexkube kubernetes client,
// which implements bunch of helper methods for Kubernetes API.
func NewClient(kubeconfig []byte) (Client, error) {
	c, err := NewClientset(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed creating kubernetes clientset: %w", err)
	}

	return &client{c}, nil
}

// PingWait waits for Kubernetes API to become available.
func (c *client) PingWait(pollInterval, retryTimeout time.Duration) error {
	return wait.PollImmediate(pollInterval, retryTimeout, c.ping)
}

// ping checks availability of Kubernetes API by fetching all Roles in kube-system namespace.
// We use Roles, as helm client sometimes fails, even if API is already available,
// saying that this type of object is not recognized.
//
//nolint:nilerr // Ignore errors here, as we only poke cluster to make sure it's functional.
func (c *client) ping() (bool, error) {
	if _, err := c.RbacV1().Roles("").List(context.TODO(), metav1.ListOptions{}); err != nil {
		return false, nil
	}

	if _, err := c.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{}); err != nil {
		return false, nil
	}

	if _, err := c.PolicyV1beta1().PodSecurityPolicies().List(context.TODO(), metav1.ListOptions{}); err != nil {
		return false, nil
	}

	return true, nil
}

// CheckNodeExists checks if given node object exists.
func (c *client) CheckNodeExists(name string) func() (bool, error) {
	return func() (bool, error) {
		_, err := c.CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}

		if errors.IsNotFound(err) {
			return false, nil
		}

		return false, fmt.Errorf("getting node %q: %w", name, err)
	}
}

// CheckNodeReady checks if given node object is ready.
func (c *client) CheckNodeReady(name string) func() (bool, error) {
	return func() (bool, error) {
		n, err := c.CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, nil //nolint:nilerr // Ignore all errors, node should eventually come up.
		}

		for _, condition := range n.Status.Conditions {
			if condition.Type == "Ready" && condition.Status != v1.ConditionTrue {
				return true, nil
			}
		}

		return false, nil
	}
}

// WaitForNode waits for node object. If object is not found and we reach the timeout, error is returned.
func (c *client) WaitForNode(name string) error {
	return wait.PollImmediate(PollInterval, RetryTimeout, c.CheckNodeExists(name))
}

// WaitForNode waits for node object to become ready. If object is not found and we reach the timeout,
// error is returned.
func (c *client) WaitForNodeReady(name string) error {
	return wait.PollImmediate(PollInterval, RetryTimeout, c.CheckNodeReady(name))
}

// LabelNode add specified labels to the Node object. If label already exist, it will be replaced.
func (c *client) LabelNode(name string, labels map[string]string) error {
	if err := c.WaitForNode(name); err != nil {
		return fmt.Errorf("failed waiting for node: %w", err)
	}

	patches := []patchStringValue{}

	for k, v := range labels {
		patches = append(patches, patchStringValue{
			Op:    "replace",
			Path:  fmt.Sprintf("/metadata/labels/%s", strings.ReplaceAll(strings.ReplaceAll(k, "~", "~0"), "/", "~1")),
			Value: v,
		})
	}

	payloadBytes, err := json.Marshal(patches)
	if err != nil {
		return fmt.Errorf("failed to encode update payload: %w", err)
	}

	if _, err := c.CoreV1().Nodes().Patch(context.TODO(), name, types.JSONPatchType, payloadBytes, metav1.PatchOptions{}); err != nil {
		return fmt.Errorf("patching node %q: %w", name, err)
	}

	return nil
}
