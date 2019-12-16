package controlplane

import (
	"testing"
)

func TestControlplaneFromYaml(t *testing.T) {
	c := `
kubernetesCACertificate: foo
frontProxyCACertificate: foo
kubeAPIServer:
  apiServerCertificate: foo
  apiServerKey: foo
  frontProxyCertificate: foo
  frontProxyKey: foo
  kubeletClientCertificate: foo
  kubeletClientKey: foo
  serviceAccountPublicKey: foo
  serviceCIDR: 11.0.0.0/24
  etcdServers:
  - http://10.0.2.15:2379
  bindAddress: 0.0.0.0
  advertiseAddress: 127.0.0.1
kubeControllerManager:
  kubernetesCAKey: foo
  serviceAccountPrivateKey: foo
  clientCertificate: foo
  clientKey: foo
  rootCACertificate: foo
  apiServer: 127.0.0.1:6443
kubeScheduler:
  clientCertificate: foo
  clientKey: foo
  apiServer: 127.0.0.1:6443
apiServerPort: 6443
ssh:
  user: "core"
  address: 127.0.0.1
  port: 2222
  password: "foo"
`

	if _, err := FromYaml([]byte(c)); err != nil {
		t.Fatalf("Creating controlplane from YAML should succeed, got: %v", err)
	}
}
