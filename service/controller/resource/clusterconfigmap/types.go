package clusterconfigmap

type ChartOperatorConfig struct {
	Cni map[string]bool `yaml:"cni"`
}
type KubernetesConfig struct {
	API map[string]string `yaml:"api"`
	DNS map[string]string `yaml:"dns"`
}
type ClusterConfig struct {
	Calico     map[string]string `yaml:"calico"`
	Kubernetes KubernetesConfig  `yaml:"kubernetes"`
}
type ClusterValuesConfig struct {
	BaseDomain string        `yaml:"baseDomain"`
	Cluster    ClusterConfig `yaml:"cluster"`
	ClusterCA  string        `yaml:"clusterCA"`
	// ClusterDNSIP is used by chart-operator. It uses this IP as its dnsConfig nameserver, to use it as resolver.
	ClusterDNSIP  string              `yaml:"clusterDNSIP"`
	ClusterID     string              `yaml:"clusterID"`
	ClusterCIDR   string              `yaml:"clusterCIDR"`
	Provider      string              `yaml:"provider"`
	GcpProject    string              `yaml:"gcpProject"`
	ChartOperator ChartOperatorConfig `yaml:"chartOperator"`
}
