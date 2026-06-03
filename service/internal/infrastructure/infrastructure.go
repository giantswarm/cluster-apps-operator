package provider

const (
	AWSClusterKind         = "AWSCluster"
	AWSClusterKindProvider = "capa"

	AWSManagedClusterKind = "AWSManagedCluster"

	AzureClusterKind         = "AzureCluster"
	AzureClusterKindProvider = "capz"

	AzureManagedClusterKind = "AzureManagedCluster"

	AzureASOManagedClusterKind      = "AzureASOManagedCluster"
	AzureASOManagedControlPlaneKind = "AzureASOManagedControlPlane"

	VCDClusterKind         = "VCDCluster"
	VCDClusterKindProvider = "cloud-director"

	VSphereClusterKind         = "VSphereCluster"
	VSphereClusterKindProvider = "vsphere"

	GCPClusterKind         = "GCPCluster"
	GCPClusterKindProvider = "gcp"

	GCPManagedClusterKind = "GCPManagedCluster"

	ProxmoxClusterKind         = "ProxmoxCluster"
	ProxmoxClusterKindProvider = "proxmox"
)
