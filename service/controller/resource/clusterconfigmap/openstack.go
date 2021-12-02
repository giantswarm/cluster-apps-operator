package clusterconfigmap

type openStackClouds struct {
	Clouds map[string]openStackCloudConfig `json:"clouds"`
}

type openStackCloudConfigAuth struct {
	AuthURL        string `json:"auth_url"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	UserDomainName string `json:"user_domain_name"`
	ProjectID      string `json:"project_id"`
}

type openStackCloudConfig struct {
	Auth               openStackCloudConfigAuth `json:"auth"`
	Verify             bool                     `json:"verify"`
	RegionName         string                   `json:"region_name"`
	Interface          string                   `json:"interface"`
	IdentityAPIVersion int                      `json:"identity_api_version"`
}
