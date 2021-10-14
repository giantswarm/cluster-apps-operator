module github.com/giantswarm/cluster-apps-operator

go 1.16

require (
	github.com/giantswarm/apiextensions/v3 v3.34.0
	github.com/giantswarm/apptest v0.12.0
	github.com/giantswarm/backoff v0.2.0
	github.com/giantswarm/exporterkit v0.2.1
	github.com/giantswarm/k8sclient/v5 v5.12.0
	github.com/giantswarm/k8smetadata v0.4.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/microkit v0.2.2
	github.com/giantswarm/micrologger v0.5.0
	github.com/giantswarm/operatorkit/v5 v5.0.0
	github.com/giantswarm/resource/v3 v3.0.2
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/viper v1.8.1
	k8s.io/api v0.20.11
	k8s.io/apimachinery v0.20.11
	k8s.io/client-go v0.20.11
	sigs.k8s.io/cluster-api v0.4.4
	sigs.k8s.io/cluster-api-provider-azure v0.5.3
	sigs.k8s.io/controller-runtime v0.6.5
	sigs.k8s.io/yaml v1.3.0
)

replace (
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	// Use go-logr/logr v0.1.0 due to breaking changes in v0.2.0 that can't be applied.
	github.com/go-logr/logr v0.2.0 => github.com/go-logr/logr v0.1.0
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
	// Same as go-logr/logr, klog/v2 is using logr v0.2.0
	k8s.io/klog/v2 v2.4.0 => k8s.io/klog/v2 v2.0.0
	sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.13-gs
	sigs.k8s.io/cluster-api-provider-azure => github.com/giantswarm/cluster-api-provider-azure v0.4.12-gsalpha3
)
