package clusterconfigmap

type ChartOperatorBootstrapMode struct {
	Enabled          bool  `json:"enabled"`
	ApiServerPodPort int32 `json:"apiServerPodPort"`
}
type ChartOperatorConfig struct {
	Cni map[string]bool `json:"cni"`
}
type KubernetesConfig struct {
	API map[string]string `json:"API"`
	DNS map[string]string `json:"DNS"`
}
type ClusterConfig struct {
	Calico     map[string]string `json:"calico"`
	Kubernetes KubernetesConfig  `json:"kubernetes"`
	Private    bool              `json:"private"`
}
type ClusterValuesConfig struct {
	Apps       map[string]interface{} `json:"apps"`
	BaseDomain string                 `json:"baseDomain"`
	// BootstrapMode allows to configure chart-operator in bootstrap mode so that it can install charts without cni or kube-proxy.
	BootstrapMode ChartOperatorBootstrapMode `json:"bootstrapMode"`
	Cluster       ClusterConfig              `json:"cluster"`
	ClusterCA     string                     `json:"clusterCA"`
	// ClusterDNSIP is used by chart-operator. It uses this IP as its dnsConfig nameserver, to use it as resolver.
	ClusterDNSIP  string              `json:"clusterDNSIP"`
	ClusterID     string              `json:"clusterID"`
	ClusterCIDR   string              `json:"clusterCIDR"`
	ExternalDNSIP *string             `json:"externalDNSIP,omitempty"`
	Provider      string              `json:"provider"`
	GcpProject    string              `json:"gcpProject"`
	ChartOperator ChartOperatorConfig `json:"chartOperator"`
}
