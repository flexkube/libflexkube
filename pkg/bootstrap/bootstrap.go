package bootstrap

// bootstrap start static control plane on the kubelet using 'controlplane' package,
// deploys kubernetes component to it and shuts down static control plane, when self-hosted
// is up
//
// It glues following components:
// - controlplane
// - components/kubernetes
