# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.15.0] - 2023-11-10

### Changed

- Add a switch for PSP CR installation.
- Use `operatorkit` release `v7.1.0`.

## [2.14.2] - 2023-09-26

### Changed

- Bump `chart-operator` to version `v2.35.2`.

## [2.14.1] - 2023-09-21

### Changed

- Bump `chart-operator` to version `v2.35.1`.

## [2.14.0] - 2023-09-11

### Added

- Inject proxy configuration also as environment variables.
- Add proxy variables specific to cert-manager in the cluster secret.

## [2.13.0] - 2023-07-31

### Added

- Add Service Monitor.

## [2.12.0] - 2023-07-04

## [2.12.0] - 2023-07-04

### Changed

- Updated default `securityContext` values to comply with PSS policies.

### Added

- Run `chart-operator` in non-bootstrap mode for EKS.

## [2.11.1] - 2023-06-28

### Added

- Add RBAC for EKS CRs.

### Removed

- Stop pushing to `openstack-app-collection`.

## [2.11.0] - 2023-04-20

## [2.10.0] - 2023-04-20

### Added

- Detect private `capz` cluster from `AzureCluster` spec and configure chart operator accordingly
- Add push of releases to `capz-app-collection`

## [2.9.0] - 2023-04-17

### Changed

- Migrate CAPI CRDs from `v1alpha4` to `v1beta1`.

## [2.8.5] - 2023-04-05

### Changed

- Bump App Operator version to `v6.6.4`

## [2.8.4] - 2023-03-10

### Changed

- Bump App Operator version to `v6.6.2`

## [2.8.3] - 2023-03-09

### Changed

- Enable private clusters for cloud-director, openstack and vsphere.
- Added the use of the runtime/default seccomp profile.
- Add empty case for `capz` in the configmap special handling to stop error in the logs
- Bump App Operator version to `v6.6.1`

## [2.8.2] - 2022-11-29

### Changed

- Bump app-operator to 6.4.4

## [2.8.1] - 2022-11-22

### Changed

- Bump app-operator to 6.4.2

## [2.8.0] - 2022-11-18

### Changed

- `secret/cluster-values` will now be generated for all kind of providers.

## [2.7.0] - 2022-11-17

### Changed

- Bumping `chart-operator` to the `v2.33.0` version.
- `secret/cluster-values` will be now generated for `capa`.
- Configure `chart-operator` to run in private cloud enviroment withou direct direct internet access.


## [2.5.0] - 2022-11-10

### Changed

- Update api schema for CAPVCD.
- Bump app-operator to 6.4.1
- Change how Flux managed Apps are detected in the cluster deletion logic. Instead of looking at not enforced
  `giantswarm.io/managed-by` label set to `flux` we check for the existence of two common Flux labels:
  `kustomize.toolkit.fluxcd.io/name` and `kustomize.toolkit.fluxcd.io/namespace` regardless of values.

### Added

- Generating proxy-configuration for workload clusters.
  By defining a `proxy` configuration (`noProxy`,`httpProxy` and `httpsProxy`) in `configmap/cluster-apps-operator`, these information will be propagated into the cluster specific `configmap` and `secret`.
  The `noProxy` value will be computed on a cluster-base as some parameters (e.g. `baseDomain` or some defined `CIDRs` might differ).
  Apps like `cert-manager` or `chart-operator` are able to use the global configuration.

## [2.4.0] - 2022-10-17

### Changed
- Enable cluster-values secret creation for CAPVCD.

## [2.3.0] - 2022-10-10

### Changed

- Deploy `chart-operator` for workload clusters with enabled `bootstrapMode`.

## [2.2.0] - 2022-09-30

### Added

- Support for App bundles in default apps.

### Fixed

- Change default CNI subnet to not contain the mask.

## [2.1.0] - 2022-09-23

### Changed

- Bumping `chart-operator` to the `v2.30.0` version, and `app-operator` to the `v6.4.0` version.

## [2.0.3] - 2022-09-14

### Fixed

- Use right capital letters in cluster values configmap.

## [2.0.2] - 2022-09-12

### Fixed

- Use `json` marshaller instead of `yaml` to get lowercase field names in configmaps.

## [2.0.1] - 2022-09-12

### Fixed

- Add RBAC rules to get `KubeadmControlPlanes` and `GCPClusters`.

## [2.0.0] - 2022-09-12

### Added

- Add additional information from `GCPCluster` to `cluster-values` configmap when running on `gcp`.

### Changed

- The DNS IP needs to come from the cluster `Services` CIDR.
- Default `Services` CIDR changed from `172.31.0.0/16` to `10.96.0.0/12` to match k8s default.

## [1.10.0] - 2022-08-09

### Changed

- Bumping `chart-operator` to the `v2.28.0` version.

## [1.9.2] - 2022-08-08

## [1.9.1] - 2022-08-05

## [1.9.0] - 2022-08-03

### Changed

- Bumping `app-operator` to the `v6.3.0` version.
- Bumping `chart-operator` to the `v2.27.0` version.

## [1.8.2] - 2022-06-17

## [1.8.1] - 2022-06-17

### Changed

- Bumping `chart-operator` to the `v2.24.0` version.

## [1.8.0] - 2022-06-09

### Added

- Removing App CRs labeled with the `giantswarm.io/cluster` label whem removing the cluster.

## [1.7.3] - 2022-06-03

### Changed

- Temporarily ignoring nancy vulnerabilities

## [1.7.2] - 2022-06-03

### Changed

- Bump `chart-operator` version to `v2.22.0`

## [1.7.1] - 2022-05-18

### Changed

- Bump `app-operator` version to `v5.10.2`.

## [1.7.0] - 2022-05-18

### Changed

- Bump `app-operator` version to `v5.10.1`.

## [1.6.0] - 2022-05-16

### Added

- Add to `gcp-app-collection`.
- Add to `vsphere-app-collection`.

### Changed

- Bump `chart-operator` version to `v2.20.1`.
- Bump `app-operator` version to `v5.10.0`.

## [1.5.0] - 2022-03-25

### Added

- Add support for `GCPCluster`.

## [1.4.6] - 2022-03-14

### Fixed

- Fix backoff when waiting for app CRs to be deleted.

## [1.4.5] - 2022-03-11

### Changed

- Bump `app-operator` version to `v5.8.0`.

## [1.4.4] - 2022-03-09

### Changed

- Add operator to `aws-app-collection`.

## [1.4.3] - 2022-03-04

### Changed

- Move `clusterCA` to match the location expected by `dex-app`.

## [1.4.2] - 2022-03-01

### Changed

- Bump `app-operator` version to `v5.7.5`.

## [1.4.1] - 2022-03-01

### Changed

- Rename helm chart value `base` to `baseDomain` to improve clarity.
- Bump `app-operator` version to `v5.7.3`.

### Added

- Add workload cluster Kubernetes API CA to cluster values ConfigMap to support Dex configuration for OIDC.

## [1.4.0] - 2022-02-18

### Changed

- Bump app-operator to `v5.7.0`

## [1.3.0] - 2022-02-08

### Added

- Add the `cluster_apps_operator_cluster_dangling_apps` metric for detecting not yet deleted apps.

### Changed

- Bump version of the Operatorkit to `v7.0.0`.

### Fixed

- Update app-operator to `v5.6.0`.

## [1.2.0] - 2022-01-21

### Changed

- App CRs and related resources are created in the organization namespace.
- App CRs are selected using the `giantswarm.io/cluster` label.
- `app-operator` and `chart-operator` app CRs are managed by the operator and
versioned via the operator's configmap.

### Removed

- Workload cluster namespace in the management cluster is no longer created.
- Release CRs are no longer used.

## [1.1.0] - 2021-12-08

### Fixed

- Fix RBAC permissions for creating secrets and getting OpenStack clusters.

### Changed

- Adjust label selector to watch all Clusters with any `cluster-apps-operator.giantswarm.io/watching`
  label instead of those matching the current operator version to allow the operator to be
  deployed in the app collection instead of by `release-operator`.

## [1.0.0] - 2021-12-03

### Added

- Add `config.giantswarm.io/version` annotation to Chart.yaml for config
  management.
- Add support for OpenStack clusters.

### Changed

- Drop `apiextensions` dependency.
- Watch Cluster `v1alpha4` instead of `v1alpha3` (breaking change).
- Update CAPZ types to `v1alpha4`.
- Use `<cluster>.<base domain>` instead of `<cluster>.k8s.<base domain>` for cluster configmap helm template values (breaking change).

## [0.6.1] - 2021-09-13

### Fixed

- Keep cluster CR finalizer until app CRs have been deleted.

## [0.6.0] - 2021-08-31

- Add provider to the cluster values configmap.

## [0.5.0] - 2021-08-30

### Added

- Add Cluster CIDR to the cluster values configmap.

## [0.4.0] - 2021-08-26

### Changed

- Don't create AWS CNI and CoreDNS apps for EKS clusters.

### Fixed

- Fix app-admission-controller webhook name in validation error matchers.

## [0.3.1] - 2021-08-26

### Fixed

- Use `app-operator-konfigure` configmap for the app-operator per workload
cluster.

## [0.3.0] - 2021-08-23

### Added

- Check upstream CAPI cluster name label as well as Giant Swarm label.

### Fixed

- Don't remove App CR finalizer if it has not been deleted.

## [0.2.0] - 2021-08-09

### Added

- Use VPA to manage deployment resources.

## [0.1.0] - 2021-06-04

### Added

- Initial version based on app related logic extracted from cluster-operator.

[Unreleased]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.15.0...HEAD
[2.15.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.14.2...v2.15.0
[2.14.2]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.14.1...v2.14.2
[2.14.1]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.14.0...v2.14.1
[2.14.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.13.0...v2.14.0
[2.13.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.12.0...v2.13.0
[2.12.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.12.0...v2.12.0
[2.12.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.11.1...v2.12.0
[2.11.1]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.11.0...v2.11.1
[2.11.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.10.0...v2.11.0
[2.10.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.9.0...v2.10.0
[2.9.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.8.5...v2.9.0
[2.8.5]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.8.4...v2.8.5
[2.8.4]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.8.3...v2.8.4
[2.8.3]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.8.2...v2.8.3
[2.8.2]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.8.1...v2.8.2
[2.8.1]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.8.0...v2.8.1
[2.8.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.7.0...v2.8.0
[2.7.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.6.0...v2.7.0
[2.6.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.5.0...v2.6.0
[2.5.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.4.0...v2.5.0
[2.4.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.3.0...v2.4.0
[2.3.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.2.0...v2.3.0
[2.2.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.1.0...v2.2.0
[2.1.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.0.3...v2.1.0
[2.0.3]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.0.2...v2.0.3
[2.0.2]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.0.1...v2.0.2
[2.0.1]: https://github.com/giantswarm/cluster-apps-operator/compare/v2.0.0...v2.0.1
[2.0.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.10.0...v2.0.0
[1.10.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.9.2...v1.10.0
[1.9.2]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.9.1...v1.9.2
[1.9.1]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.9.0...v1.9.1
[1.9.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.8.2...v1.9.0
[1.8.2]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.8.1...v1.8.2
[1.8.1]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.8.0...v1.8.1
[1.8.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.7.3...v1.8.0
[1.7.3]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.7.2...v1.7.3
[1.7.2]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.7.1...v1.7.2
[1.7.1]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.7.0...v1.7.1
[1.7.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.6.0...v1.7.0
[1.6.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.5.0...v1.6.0
[1.5.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.4.6...v1.5.0
[1.4.6]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.4.5...v1.4.6
[1.4.5]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.4.4...v1.4.5
[1.4.4]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.4.3...v1.4.4
[1.4.3]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.4.2...v1.4.3
[1.4.2]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.4.1...v1.4.2
[1.4.1]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.4.0...v1.4.1
[1.4.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v0.6.1...v1.0.0
[0.6.1]: https://github.com/giantswarm/cluster-apps-operator/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/giantswarm/cluster-apps-operator/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/giantswarm/cluster-apps-operator/releases/tag/v0.1.0
