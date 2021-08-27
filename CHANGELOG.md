# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Add Cluster CIDR in the configmap.

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

[Unreleased]: https://github.com/giantswarm/cluster-apps-operator/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/giantswarm/cluster-apps-operator/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/giantswarm/cluster-apps-operator/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/giantswarm/cluster-apps-operator/releases/tag/v0.1.0
