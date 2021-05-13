package app

type appConfig struct {
	Catalog string `json:"catalog"`
	Version string `json:"version"`
}

type userOverrideConfig map[string]appConfig
