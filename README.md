[![CircleCI](https://circleci.com/gh/giantswarm/cluster-apps-operator.svg?&style=shield)](https://circleci.com/gh/giantswarm/cluster-apps-operator)

# cluster-apps-operator

The cluster-apps-operator is part of the Giant Swarm [App Platform].
It watches [Cluster API] v1alpha4 CRs and manages [app CRs] for workload
clusters as defined in the [release CR].

It is implemented using [operatorkit].

## Getting Project

Clone the git repository: https://github.com/giantswarm/cluster-apps-operator.git

### How to build

Build it using the standard `go build` command.

```
go build github.com/giantswarm/cluster-apps-operator
```

## Contact

- Mailing list: [giantswarm](https://groups.google.com/forum/!forum/giantswarm)
- Bugs: [issues](https://github.com/giantswarm/cluster-apps-operator/issues)

## License

cluster-apps-operator is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for
details.

[App Platform]: https://docs.giantswarm.io/app-platform/
[Cluster API]: https://cluster-api.sigs.k8s.io/
[app CRs]: https://docs.giantswarm.io/ui-api/management-api/crd/apps.application.giantswarm.io/
[release CR]: https://docs.giantswarm.io/ui-api/management-api/crd/releases.release.giantswarm.io/
[operatorkit]: https://github.com/giantswarm/operatorkit
