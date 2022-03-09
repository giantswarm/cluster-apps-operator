package project

var (
	description = "The cluster-apps-operator does something."
	gitSHA      = "n/a"
	name        = "cluster-apps-operator"
	source      = "https://github.com/giantswarm/cluster-apps-operator"
	version     = "1.4.5-dev"
)

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}
