package clustersecret

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func Test_GetProxyEnabledValueFromConfigMap(t *testing.T) {
	testCases := []struct {
		name        string
		configMap   corev1.ConfigMap
		result      bool
		expectError bool
	}{
		{
			name: "case 0",
			configMap: corev1.ConfigMap{
				Data: map[string]string{
					"values": getValuesProxyEnabled(),
				},
			},
			result: true,
		},
		{
			name: "case 1",
			configMap: corev1.ConfigMap{
				Data: map[string]string{
					"values": getValuesProxyDisabled(),
				},
			},
			result: false,
		},
		{
			name: "case 2",
			configMap: corev1.ConfigMap{
				Data: map[string]string{
					"values": getValuesProxyNotDefined(),
				},
			},
			result: false,
		},
		{
			name: "case 3",
			configMap: corev1.ConfigMap{
				Data: map[string]string{
					"values": getValuesProxyEmpty(),
				},
			},
			result: false,
		},
		{
			name: "case 4",
			configMap: corev1.ConfigMap{
				Data: map[string]string{
					"values": getValuesEmptyString(),
				},
			},
			result: false,
		},
		{
			name: "case 5",
			configMap: corev1.ConfigMap{
				Data: map[string]string{
					"values": getValuesRandomContent(),
				},
			},
			result:      false,
			expectError: true,
		},
		{
			name: "case 6",
			configMap: corev1.ConfigMap{
				Data: map[string]string{
					"wrongKey": getValuesProxyEnabled(),
				},
			},
			result: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := getProxyEnabledValueFromConfigMap(tc.configMap)
			if err != nil && !tc.expectError {
				t.Fatalf("unexpected error: %v", err)
			}
			if err == nil && tc.expectError {
				t.Fatalf("expected error, got nil")
			}
			if result != tc.result {
				t.Fatalf("result == %#v, want %#v", result, tc.result)
			}
		})
	}
}

func getValuesProxyEnabled() string {
	return `|
    global:
      release:
        version: 1.2.3
      podSecurityStandards:
        enforced: true
      connectivity:
        baseDomain: test.example.io
        proxy:
          enabled: true
          httpProxy: http://proxy.example.io:3128
          httpsProxy: http://proxy.example.io:3128
          noProxy: localhost,example.io
        availabilityZoneUsageLimit: 3
      providerSpecific:
        region: far-away-4`
}

func getValuesProxyDisabled() string {
	return `|
    global:
      release:
        version: 1.2.3
      podSecurityStandards:
        enforced: true
      connectivity:
        baseDomain: test.example.io
        proxy:
          enabled: false
        availabilityZoneUsageLimit: 3
      providerSpecific:
        region: far-away-4`
}

func getValuesProxyNotDefined() string {
	return `|
    global:
      release:
        version: 1.2.3
      podSecurityStandards:
        enforced: true
      connectivity:
        baseDomain: test.example.io
        availabilityZoneUsageLimit: 3
      providerSpecific:
        region: far-away-4`
}

func getValuesProxyEmpty() string {
	return `|
    global:
      release:
        version: 1.2.3
      podSecurityStandards:
        enforced: true
      connectivity:
        baseDomain: test.example.io
        proxy:
          enabled:
        availabilityZoneUsageLimit: 3
      providerSpecific:
        region: far-away-4`
}

func getValuesEmptyString() string {
	return `|`
}

func getValuesRandomContent() string {
	return `abcd
    efgh
    ijkl`
}
