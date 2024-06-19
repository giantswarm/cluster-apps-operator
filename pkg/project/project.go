package project

var (
	description = "The cluster-apps-operator does something."
	gitSHA      = "n/a"
	name        = "cluster-apps-operator"
	source      = "https://github.com/giantswarm/cluster-apps-operator"
	version     = "2.22.1"
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
