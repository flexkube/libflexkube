package kubelet

// Pool represents group of kubelet instances and their configuration
type Pool struct {
	Kubelets []Instance
}
