package chartname

type catalogIndex struct {
	Entries map[string][]indexEntry `json:"entries"`
}

type indexEntry struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
