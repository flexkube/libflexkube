// +build e2e

package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/cli/flexkube"
	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/pkg/apiloadbalancer"
	"github.com/flexkube/libflexkube/pkg/controlplane"
	"github.com/flexkube/libflexkube/pkg/etcd"
	"github.com/flexkube/libflexkube/pkg/helm/release"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
	"github.com/flexkube/libflexkube/pkg/kubelet"
	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
	"github.com/flexkube/libflexkube/pkg/pki"
)

type chart struct {
	Source  string `json:"source"`
	Version string `json:"version"`
}

type charts struct {
	KubeAPIServer      chart `json:"kubeAPIServer"`
	Kubernetes         chart `json:"kubernetes"`
	KubeProxy          chart `json:"kubeProxy"`
	TLSBootstrapping   chart `json:"tlsBootstrapping"`
	KubeletRubberStamp chart `json:"kubeletRubberStamp"`
	Calico             chart `json:"calico"`
	MetricsServer      chart `json:"metricsServer"`
	CoreDNS            chart `json:"coreDNS"`
}

type e2eConfig struct {
	ControllersCount  int    `json:"controllersCount"`
	NodesCIDR         string `json:"nodesCIDR"`
	FlatcarChannel    string `json:"flatcarChannel"`
	WorkersCount      int    `json:"workersCount"`
	APIPort           int    `json:"apiPort"`
	NodeSSHPort       int    `json:"nodeSSHPort"`
	SSHPrivateKeyPath string `json:"sshPrivatekeyPath"`
	Charts            charts `json:"charts"`
}

func parseInt(t *testing.T, envVar string, defaultValue int) int {
	iRaw := util.PickString(os.Getenv(envVar), fmt.Sprintf("%d", defaultValue))

	i, err := strconv.Atoi(iRaw)
	if err != nil {
		t.Fatalf("parsing %q with value %q to int: %v", envVar, iRaw, err)
	}

	return i
}

func absPath(t *testing.T, path string) string {
	p, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("Converting path %q to absolute: %v", path, err)
	}

	return p
}

func defaultE2EConfig(t *testing.T) e2eConfig {
	return e2eConfig{
		ControllersCount:  parseInt(t, "TF_VAR_controllers_count", 1),
		WorkersCount:      parseInt(t, "TF_VAR_workers_count", 0),
		NodesCIDR:         util.PickString(os.Getenv("TF_VAR_nodes_cidr"), "192.168.50.0/24"),
		FlatcarChannel:    util.PickString(os.Getenv("TF_VAR_flatcar_channel"), "stable"),
		APIPort:           8443,
		NodeSSHPort:       22,
		SSHPrivateKeyPath: "/root/.ssh/id_rsa",
		Charts: charts{
			KubeAPIServer: chart{
				Source:  "flexkube/kube-apiserver",
				Version: "0.2.1",
			},
			Kubernetes: chart{
				Source:  "flexkube/kubernetes",
				Version: "0.3.3",
			},
			KubeProxy: chart{
				Source:  "flexkube/kube-proxy",
				Version: "0.2.4",
			},
			TLSBootstrapping: chart{
				Source:  "flexkube/tls-bootstrapping",
				Version: "0.1.1",
			},
			CoreDNS: chart{
				Source:  "stable/coredns",
				Version: "1.13.3",
			},
			MetricsServer: chart{
				Source:  "stable/metrics-server",
				Version: "2.11.1",
			},
			KubeletRubberStamp: chart{
				Source:  "flexkube/kubelet-rubber-stamp",
				Version: "0.1.4",
			},
			Calico: chart{
				Source:  "flexkube/calico",
				Version: "0.2.5",
			},
		},
	}
}

//nolint:funlen,gocognit,gocyclo
func TestE2e(t *testing.T) {
	testConfig := defaultE2EConfig(t)

	testConfigFile := "test-config.yaml"

	tc, err := readYamlFile(testConfigFile)
	if err != nil {
		t.Fatalf("Reading test config file %q: %v", testConfigFile, err)
	}

	if err := yaml.Unmarshal(tc, &testConfig); err != nil {
		t.Fatalf("Parsing test config file %q: %v", testConfigFile, err)
	}

	t.Logf("Running with following configuration: \n%s\n", cmp.Diff("", testConfig))

	ip, ipnet, err := net.ParseCIDR(testConfig.NodesCIDR)
	if err != nil {
		t.Fatalf("parsing nodes CIDR %q: %v", testConfig.NodesCIDR, err)
	}

	// Calculate controllers IPs and names.
	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	peers := map[string]string{}
	controllerIPs := []string{}
	controllerNames := []string{}
	servers := []string{}
	members := map[string]etcd.Member{}
	controllerLBs := []apiloadbalancer.APILoadBalancer{}
	etcdServers := []string{}
	controllerKubelets := []kubelet.Kubelet{}
	workerLBs := []apiloadbalancer.APILoadBalancer{}
	workerKubelets := []kubelet.Kubelet{}

	for i := 0; i < testConfig.WorkersCount; i++ {
		name := fmt.Sprintf("worker%02d", i+1)
		ip := ips[i+2+10]

		host := host.Host{
			SSHConfig: &ssh.Config{
				Address: ip,
			},
		}

		workerLBs = append(workerLBs, apiloadbalancer.APILoadBalancer{
			Host: host,
		})

		workerKubelets = append(workerKubelets, kubelet.Kubelet{
			Name:    name,
			Address: ip,
			Host:    host,
		})
	}

	for i := 0; i < testConfig.ControllersCount; i++ {
		name := fmt.Sprintf("controller%02d", i+1)
		ip := ips[i+2]

		controllerNames = append(controllerNames, name)
		controllerIPs = append(controllerIPs, ip)
		peers[name] = ip
		servers = append(servers, fmt.Sprintf("%s:%d", ip, testConfig.APIPort))

		host := host.Host{
			SSHConfig: &ssh.Config{
				Address: ip,
			},
		}

		members[name] = etcd.Member{
			Name:          name,
			PeerAddress:   ip,
			ServerAddress: ip,
			Host:          host,
		}

		controllerLBs = append(controllerLBs, apiloadbalancer.APILoadBalancer{
			Host: host,
		})

		etcdServers = append(etcdServers, fmt.Sprintf("https://%s:2379", ip))

		controllerKubelets = append(controllerKubelets, kubelet.Kubelet{
			Name:    name,
			Address: ip,
			Host:    host,
		})
	}

	t.Logf("Controller IPs: %s", strings.Join(controllerIPs, ", "))
	t.Logf("Controller names: %s", strings.Join(controllerNames, ", "))

	nodeSSHPort := testConfig.NodeSSHPort

	sshPrivateKey, err := ioutil.ReadFile(testConfig.SSHPrivateKeyPath)
	if err != nil {
		t.Fatalf("Reading SSH private key %q: %v", testConfig.SSHPrivateKeyPath, err)
	}

	sshUser := "core"

	sshConfig := &ssh.Config{
		User:       sshUser,
		Port:       nodeSSHPort,
		PrivateKey: string(sshPrivateKey),
	}

	// Static bootstrap token, so it does not get changed on every test run.
	bootstrapTokenID := "64vxqx"
	bootstrapTokenSecret := "z95f5ng9sek5i40v" // #nosec:G101

	cgroupDriver := "cgroupfs"
	if testConfig.FlatcarChannel == "edge" {
		cgroupDriver = "systemd"
	}

	t.Logf("Using cgroup driver: %s", cgroupDriver)

	networkPlugin := "cni"
	hairpinMode := "hairpin-veth"

	// Generate PKI.
	r := &flexkube.Resource{
		Confirmed: true,
		PKI: &pki.PKI{
			Etcd: &pki.Etcd{
				Peers:   peers,
				Servers: peers,
				ClientCNs: []string{
					"root",
					"kube-apiserver",
					"prometheus",
				},
			},
			Kubernetes: &pki.Kubernetes{
				KubeAPIServer: &pki.KubeAPIServer{
					ExternalNames: []string{"kube-apiserver.example.com"},
					ServerIPs:     append(controllerIPs, "127.0.1.1", "11.0.0.1"),
				},
			},
		},
		Etcd: &etcd.Cluster{
			SSH:     sshConfig,
			Members: members,
		},
		APILoadBalancerPools: map[string]*apiloadbalancer.APILoadBalancers{
			"controllers": {
				Name:             "api-loadbalancer-controllers",
				HostConfigPath:   "/etc/haproxy/controllers.cfg",
				BindAddress:      "127.0.0.1:7443",
				Servers:          servers,
				SSH:              sshConfig,
				APILoadBalancers: controllerLBs,
			},
		},
		Controlplane: &controlplane.Controlplane{
			KubeAPIServer: controlplane.KubeAPIServer{
				ServiceCIDR:      "11.0.0.0/24",
				EtcdServers:      etcdServers,
				BindAddress:      controllerIPs[0],
				AdvertiseAddress: controllerIPs[0],
				SecurePort:       testConfig.APIPort,
			},
			KubeControllerManager: controlplane.KubeControllerManager{
				FlexVolumePluginDir: "/var/lib/kubelet/volumeplugins",
			},
			APIServerPort:    testConfig.APIPort,
			APIServerAddress: controllerIPs[0],
			SSH: &ssh.Config{
				User:       sshUser,
				Port:       nodeSSHPort,
				PrivateKey: string(sshPrivateKey),
				Address:    controllerIPs[0],
			},
		},
		KubeletPools: map[string]*kubelet.Pool{
			"controller": {
				BootstrapConfig: &client.Config{
					Server: "127.0.0.1:7443",
					Token:  fmt.Sprintf("%s.%s", bootstrapTokenID, bootstrapTokenSecret),
				},
				WaitForNodeReady: true,
				CgroupDriver:     cgroupDriver,
				NetworkPlugin:    networkPlugin,
				HairpinMode:      hairpinMode,
				VolumePluginDir:  "/var/lib/kubelet/volumeplugins",
				ClusterDNSIPs:    []string{"11.0.0.10"},
				SystemReserved: map[string]string{
					"cpu":    "100m",
					"memory": "500Mi",
				},
				KubeReserved: map[string]string{
					"cpu": "100m",
					// 100MB for kubelet and 200MB for etcd.
					"memory": "300Mi",
				},
				PrivilegedLabels: map[string]string{
					"node-role.kubernetes.io/master": "",
				},
				AdminConfig: &client.Config{
					Server: fmt.Sprintf("%s:%d", controllerIPs[0], testConfig.APIPort),
				},
				Taints: map[string]string{
					"node-role.kubernetes.io/master": "NoSchedule",
				},
				SSH:      sshConfig,
				Kubelets: controllerKubelets,
			},
		},
		State: &flexkube.ResourceState{},
	}

	if testConfig.WorkersCount > 0 {
		r.KubeletPools["workers"] = &kubelet.Pool{
			BootstrapConfig: &client.Config{
				Server: "127.0.0.1:7443",
				Token:  fmt.Sprintf("%s.%s", bootstrapTokenID, bootstrapTokenSecret),
			},
			WaitForNodeReady: true,
			CgroupDriver:     cgroupDriver,
			NetworkPlugin:    networkPlugin,
			HairpinMode:      hairpinMode,
			VolumePluginDir:  "/var/lib/kubelet/volumeplugins",
			ClusterDNSIPs:    []string{"11.0.0.10"},
			SystemReserved: map[string]string{
				"cpu":    "100m",
				"memory": "500Mi",
			},
			KubeReserved: map[string]string{
				"cpu":    "100m",
				"memory": "100Mi",
			},
			AdminConfig: &client.Config{
				Server: fmt.Sprintf("%s:%d", controllerIPs[0], testConfig.APIPort),
			},
			SSH:      sshConfig,
			Kubelets: workerKubelets,
		}

		r.APILoadBalancerPools["workers"] = &apiloadbalancer.APILoadBalancers{
			Name:             "api-loadbalancer-workers",
			HostConfigPath:   "/etc/haproxy/workers.cfg",
			BindAddress:      "127.0.0.1:7443",
			Servers:          servers,
			SSH:              sshConfig,
			APILoadBalancers: workerLBs,
		}
	}

	resourceRaw, err := yaml.Marshal(r)
	if err != nil {
		t.Fatalf("Serializing resource configuration: %v", err)
	}

	if err := ioutil.WriteFile("config.yaml", resourceRaw, 0o600); err != nil {
		t.Fatalf("Writing config.yaml file: %v", err)
	}

	// Read state.
	resourceStateFile := "state.yaml"

	s, err := readYamlFile(resourceStateFile)
	if err != nil {
		t.Fatalf("Reading state file %q: %v", resourceStateFile, err)
	}

	if err := yaml.Unmarshal(s, r); err != nil {
		t.Fatalf("Loading PKI state failed: %v", err)
	}

	// Deploy things.
	if err := r.StateToFile(r.RunPKI()); err != nil {
		t.Fatalf("Running PKI: %v", err)
	}

	if err := r.StateToFile(r.RunEtcd()); err != nil {
		t.Fatalf("Running etcd: %v", err)
	}

	for k := range r.APILoadBalancerPools {
		if err := r.StateToFile(r.RunAPILoadBalancerPool(k)); err != nil {
			t.Fatalf("Running API load balancer pool %q: %v", k, err)
		}
	}

	if err := r.StateToFile(r.RunControlplane()); err != nil {
		t.Fatalf("Running controlplane: %v", err)
	}

	// Kubeconfig.
	k, err := r.Kubeconfig()
	if err != nil {
		t.Fatalf("Getting kubeconfig: %v", err)
	}

	for _, dir := range []string{"./resources/etcd-cluster", "values"} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("Creating directory %q: %v", dir, err)
		}
	}

	etcdTemplate := `#!/bin/bash
export ETCDCTL_API=3
export ETCDCTL_CACERT=%s
export ETCDCTL_CERT=%s
export ETCDCTL_KEY=%s
export ETCDCTL_ENDPOINTS=%s
`
	files := map[string]string{
		"kubeconfig":                                     k,
		"./resources/etcd-cluster/ca.pem":                string(r.State.PKI.Etcd.CA.X509Certificate),
		"./resources/etcd-cluster/client.pem":            string(r.State.PKI.Etcd.ClientCertificates["root"].X509Certificate),
		"./resources/etcd-cluster/client.key":            string(r.State.PKI.Etcd.ClientCertificates["root"].PrivateKey),
		"./resources/etcd-cluster/prometheus_client.pem": string(r.State.PKI.Etcd.ClientCertificates["prometheus"].X509Certificate),
		"./resources/etcd-cluster/prometheus_client.key": string(r.State.PKI.Etcd.ClientCertificates["prometheus"].PrivateKey),
		"./resources/etcd-cluster/environment.sh": fmt.Sprintf(etcdTemplate,
			absPath(t, "./resources/etcd-cluster/ca.pem"),
			absPath(t, "./resources/etcd-cluster/client.pem"),
			absPath(t, "./resources/etcd-cluster/client.key"),
			strings.Join(etcdServers, ","),
		),
		"./resources/etcd-cluster/prometheus-environment.sh": fmt.Sprintf(etcdTemplate,
			absPath(t, "./resources/etcd-cluster/ca.pem"),
			absPath(t, "./resources/etcd-cluster/prometheus_client.pem"),
			absPath(t, "./resources/etcd-cluster/prometheus_client.key"),
			strings.Join(etcdServers, ","),
		),
		"./resources/etcd-cluster/enable-rbac.sh": `
#!/bin/bash
etcdctl user add --no-password=true root
etcdctl role add root
etcdctl user grant-role root root
etcdctl auth enable
etcdctl user add --no-password=true kube-apiserver
etcdctl role add kube-apiserver
etcdctl role grant-permission kube-apiserver readwrite --prefix=true /
etcdctl user grant-role kube-apiserver kube-apiserver
# Until https://github.com/etcd-io/etcd/issues/8458 is resolved.
etcdctl user grant-role kube-apiserver root
etcdctl user add --no-password=true prometheus
`,
	}

	// TLS bootstrapping.
	values := fmt.Sprintf(`
tokens:
- token-id: %s
  token-secret: %s
`, bootstrapTokenID, bootstrapTokenSecret)

	files["./values/tls-bootstrapping.yaml"] = values

	config := &release.Config{
		Kubeconfig: k,
		Namespace:  "kube-system",
		Name:       "tls-bootstrapping",
		Version:    testConfig.Charts.TLSBootstrapping.Version,
		Chart:      testConfig.Charts.TLSBootstrapping.Source,
		Values:     values,
	}

	rel, err := config.New()
	if err != nil {
		t.Fatalf("Creating TLS bootstrapping release object: %v", err)
	}

	if err := rel.InstallOrUpgrade(); err != nil {
		t.Fatalf("Installing TLS bootstrapping release %q: %v", config.Name, err)
	}

	// Deploy kubelets.
	for k := range r.KubeletPools {
		if err := r.StateToFile(r.RunKubeletPool(k)); err != nil {
			t.Fatalf("Running kubelet pool %q: %v", k, err)
		}
	}

	// Kube-proxy.
	values, err = r.TemplateFromFile("./templates/kube-proxy-values.yaml.tmpl")
	if err != nil {
		t.Fatalf("Executing kube-proxy values template: %v", err)
	}

	files["./values/kube-proxy.yaml"] = values

	config = &release.Config{
		Kubeconfig: k,
		Namespace:  "kube-system",
		Name:       "kube-proxy",
		Version:    testConfig.Charts.KubeProxy.Version,
		Chart:      testConfig.Charts.KubeProxy.Source,
		Values:     values,
		Wait:       true,
	}

	rel, err = config.New()
	if err != nil {
		t.Fatalf("Creating kube-proxy release object: %v", err)
	}

	if err := rel.InstallOrUpgrade(); err != nil {
		t.Fatalf("Installing kube-proxy release %q: %v", config.Name, err)
	}

	// Calico.
	values = `
podCIDR: 10.1.0.0/16
flexVolumePluginDir: /var/lib/kubelet/volumeplugins
`

	files["./values/calico.yaml"] = values

	config = &release.Config{
		Kubeconfig: k,
		Namespace:  "kube-system",
		Name:       "calico",
		Version:    testConfig.Charts.Calico.Version,
		Chart:      testConfig.Charts.Calico.Source,
		Values:     values,
		Wait:       true,
	}

	rel, err = config.New()
	if err != nil {
		t.Fatalf("Creating TLS bootstrapping release object: %v", err)
	}

	if err := rel.InstallOrUpgrade(); err != nil {
		t.Fatalf("Installing TLS bootstrapping release %q: %v", config.Name, err)
	}

	// kube-apiserver.
	values, err = r.TemplateFromFile("./templates/kube-apiserver-values.yaml.tmpl")
	if err != nil {
		t.Fatalf("Executing template: %v", err)
	}

	files["./values/kube-apiserver.yaml"] = values

	config = &release.Config{
		Kubeconfig: k,
		Namespace:  "kube-system",
		Name:       "kube-apiserver",
		Version:    testConfig.Charts.KubeAPIServer.Version,
		Chart:      testConfig.Charts.KubeAPIServer.Source,
		Values:     values,
		Wait:       true,
	}

	rel, err = config.New()
	if err != nil {
		t.Fatalf("Creating release object: %v", err)
	}

	if err := rel.InstallOrUpgrade(); err != nil {
		t.Fatalf("Installing release %q: %v", config.Name, err)
	}

	// Kubernetes.
	values, err = r.TemplateFromFile("./templates/kubernetes-values.yaml.tmpl")
	if err != nil {
		t.Fatalf("Executing Kubernetes values template: %v", err)
	}

	files["./values/kubernetes.yaml"] = values

	config = &release.Config{
		Kubeconfig: k,
		Namespace:  "kube-system",
		Name:       "kubernetes",
		Version:    testConfig.Charts.Kubernetes.Version,
		Chart:      testConfig.Charts.Kubernetes.Source,
		Values:     values,
		Wait:       true,
	}

	rel, err = config.New()
	if err != nil {
		t.Fatalf("Creating Kubernetes release object: %v", err)
	}

	if err := rel.InstallOrUpgrade(); err != nil {
		t.Fatalf("Installing Kubernetes release %q: %v", config.Name, err)
	}

	// CoreDNS.
	values = `
rbac:
  pspEnable: true
service:
  clusterIP: 11.0.0.10
nodeSelector:
  node-role.kubernetes.io/master: ""
tolerations:
  - key: node-role.kubernetes.io/master
    operator: Exists
    effect: NoSchedule
`
	files["./values/coredns.yaml"] = values

	config = &release.Config{
		Kubeconfig: k,
		Namespace:  "kube-system",
		Name:       "coredns",
		Version:    testConfig.Charts.CoreDNS.Version,
		Chart:      testConfig.Charts.CoreDNS.Source,
		Values:     values,
		Wait:       true,
	}

	rel, err = config.New()
	if err != nil {
		t.Fatalf("Creating CoreDNS release object: %v", err)
	}

	if err := rel.InstallOrUpgrade(); err != nil {
		t.Fatalf("Installing CoreDNS release %q: %v", config.Name, err)
	}

	// Metrics server.
	values = `
rbac:
  pspEnabled: true
args:
- --kubelet-preferred-address-types=InternalIP
podDisruptionBudget:
  enabled: true
  minAvailable: 1
tolerations:
- key: node-role.kubernetes.io/master
  operator: Exists
  effect: NoSchedule
resources:
  requests:
    memory: 20Mi
`
	files["./values/metrics-server.yaml"] = values

	config = &release.Config{
		Kubeconfig: k,
		Namespace:  "kube-system",
		Name:       "metrics-server",
		Version:    testConfig.Charts.MetricsServer.Version,
		Chart:      testConfig.Charts.MetricsServer.Source,
		Values:     values,
		Wait:       true,
	}

	rel, err = config.New()
	if err != nil {
		t.Fatalf("Creating Metrics server release object: %v", err)
	}

	if err := rel.InstallOrUpgrade(); err != nil {
		t.Fatalf("Installing Metrics server release %q: %v", config.Name, err)
	}

	// Kubelet-rubber-stamp.
	config = &release.Config{
		Kubeconfig: k,
		Namespace:  "kube-system",
		Name:       "kubelet-rubber-stamp",
		Version:    testConfig.Charts.KubeletRubberStamp.Version,
		Chart:      testConfig.Charts.KubeletRubberStamp.Source,
		Wait:       true,
	}

	rel, err = config.New()
	if err != nil {
		t.Fatalf("Creating Kubelet rubber stamp release object: %v", err)
	}

	if err := rel.InstallOrUpgrade(); err != nil {
		t.Fatalf("Installing Kubelet rubber stamp release %q: %v", config.Name, err)
	}

	for file, content := range files {
		if err := ioutil.WriteFile(file, []byte(content), 0o600); err != nil {
			t.Fatalf("Writing file %q: %v", file, err)
		}
	}
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
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

	// Workaround for empty YAML file.
	if string(c) == "{}\n" {
		return []byte{}, nil
	}

	return c, nil
}
