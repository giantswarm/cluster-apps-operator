package integration

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/cluster-apps-operator/v3/flag/service/proxy"
	"github.com/giantswarm/cluster-apps-operator/v3/test/harness"
)

func TestClusterSecret_Basic(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()
	ns := env.CreateNamespace(ctx, "test-secret-basic-"+randomSuffix())

	env.CreateInfraCluster(ctx, schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "GCPCluster",
		Version: "v1beta1",
	}, ns.Name, "test-cluster", map[string]interface{}{
		"project": "12345",
		"region":  "us-central1",
	})

	cluster := env.CreateCluster(ctx, harness.ClusterOpts{
		Name:      "test-cluster",
		Namespace: ns.Name,
		Labels:    map[string]string{capi.ClusterNameLabel: "test-cluster"},
		InfraRef: &corev1.ObjectReference{
			Kind:       "GCPCluster",
			Namespace:  ns.Name,
			Name:       "test-cluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
	})

	resource := env.NewSecretResource(harness.SecretResourceOpts{})

	secrets, err := resource.GetDesiredState(ctx, cluster)
	if err != nil {
		t.Fatal(err)
	}

	if len(secrets) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(secrets))
	}

	secret := secrets[0]
	if !strings.HasSuffix(secret.Name, "-cluster-values") {
		t.Fatalf("expected secret name to end with -cluster-values, got %q", secret.Name)
	}

	valuesData, ok := secret.Data["values"]
	if !ok {
		t.Fatal("secret missing 'values' key")
	}

	var values map[string]interface{}
	if err := yaml.Unmarshal(valuesData, &values); err != nil {
		t.Fatal(err)
	}

	// No proxy config should be present
	if _, found := values["http_proxy"]; found {
		t.Fatal("http_proxy should not be set without proxy config")
	}
	if _, found := values["https_proxy"]; found {
		t.Fatal("https_proxy should not be set without proxy config")
	}
	if _, found := values["no_proxy"]; found {
		t.Fatal("no_proxy should not be set without proxy config")
	}
}

func TestClusterSecret_WithProxy(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()
	ns := env.CreateNamespace(ctx, "test-secret-proxy-"+randomSuffix())

	// Create AzureCluster with Internal LB so privatecluster returns true
	env.CreateInfraCluster(ctx, schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "AzureCluster",
		Version: "v1beta1",
	}, ns.Name, "test-cluster", map[string]interface{}{
		"location":       "eastus",
		"resourceGroup":  "group1",
		"subscriptionID": "sub-123",
		"networkSpec": map[string]interface{}{
			"apiServerLB": map[string]interface{}{
				"type": "Internal",
			},
		},
	})

	cluster := env.CreateCluster(ctx, harness.ClusterOpts{
		Name:      "test-cluster",
		Namespace: ns.Name,
		Labels:    map[string]string{capi.ClusterNameLabel: "test-cluster"},
		InfraRef: &corev1.ObjectReference{
			Kind:       "AzureCluster",
			Namespace:  ns.Name,
			Name:       "test-cluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
		ClusterNetwork: &capi.ClusterNetwork{
			ServiceDomain: "cluster.local",
			Services:      &capi.NetworkRanges{CIDRBlocks: []string{"172.31.0.0/16"}},
			Pods:          &capi.NetworkRanges{CIDRBlocks: []string{"192.168.0.0/16"}},
		},
	})

	resource := env.NewSecretResource(harness.SecretResourceOpts{
		Proxy: proxy.Proxy{
			HttpProxy:  "http://proxy.example.com:3128",
			HttpsProxy: "https://proxy.example.com:3128",
			NoProxy:    "10.0.0.0/8",
		},
	})

	secrets, err := resource.GetDesiredState(ctx, cluster)
	if err != nil {
		t.Fatal(err)
	}

	if len(secrets) != 2 {
		t.Fatalf("expected 2 secrets (cluster-values + systemd-proxy), got %d", len(secrets))
	}

	// Find cluster-values and systemd-proxy secrets
	var clusterValuesSecret, systemdProxySecret *corev1.Secret
	for _, s := range secrets {
		if strings.HasSuffix(s.Name, "-cluster-values") {
			clusterValuesSecret = s
		}
		if strings.HasSuffix(s.Name, "-systemd-proxy") {
			systemdProxySecret = s
		}
	}

	if clusterValuesSecret == nil {
		t.Fatal("cluster-values secret not found")
	}
	if systemdProxySecret == nil {
		t.Fatal("systemd-proxy secret not found")
	}

	// Verify cluster-values secret has proxy config
	var values map[string]interface{}
	if err := yaml.Unmarshal(clusterValuesSecret.Data["values"], &values); err != nil {
		t.Fatal(err)
	}

	if values["http_proxy"] != "http://proxy.example.com:3128" {
		t.Fatalf("expected http_proxy to be %q, got %v", "http://proxy.example.com:3128", values["http_proxy"])
	}
	if values["https_proxy"] != "https://proxy.example.com:3128" {
		t.Fatalf("expected https_proxy to be %q, got %v", "https://proxy.example.com:3128", values["https_proxy"])
	}
	noProxy, ok := values["no_proxy"].(string)
	if !ok {
		t.Fatal("no_proxy should be a string")
	}
	if !strings.Contains(noProxy, "10.0.0.0/8") {
		t.Fatalf("no_proxy should contain 10.0.0.0/8, got %q", noProxy)
	}

	// Verify cluster proxy block
	clusterVal, ok := values["cluster"].(map[string]interface{})
	if !ok {
		t.Fatal("cluster key should be a map")
	}
	proxyVal, ok := clusterVal["proxy"].(map[string]interface{})
	if !ok {
		t.Fatal("cluster.proxy key should be a map")
	}
	if proxyVal["http"] != "http://proxy.example.com:3128" {
		t.Fatalf("expected cluster.proxy.http to be %q, got %v", "http://proxy.example.com:3128", proxyVal["http"])
	}
	if proxyVal["https"] != "https://proxy.example.com:3128" {
		t.Fatalf("expected cluster.proxy.https to be %q, got %v", "https://proxy.example.com:3128", proxyVal["https"])
	}

	// Verify systemd-proxy secret has containerd proxy data
	containerdProxy, ok := systemdProxySecret.Data["containerdProxy"]
	if !ok {
		t.Fatal("systemd-proxy secret missing 'containerdProxy' key")
	}
	proxyContent := string(containerdProxy)
	if !strings.Contains(proxyContent, "http://proxy.example.com:3128") {
		t.Fatalf("containerdProxy should contain HTTP_PROXY, got %q", proxyContent)
	}
	if !strings.Contains(proxyContent, "https://proxy.example.com:3128") {
		t.Fatalf("containerdProxy should contain HTTPS_PROXY, got %q", proxyContent)
	}
}
