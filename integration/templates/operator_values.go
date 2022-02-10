package templates

// ClusterAppsOperatorValues values required by the cluster-apps-operator chart.
const ClusterAppsOperatorValues = `
 baseDomain: example.com
 provider:
   kind: aws
 registry:
   domain: quay.io
`
