# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/giantswarm/cluster-apps-operator/compare/v1.7.0...HEAD
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
