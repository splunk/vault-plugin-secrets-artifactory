package artifactorysecrets

type Permission struct {
	IncludePatterns []string `json:"include_patterns,omitempty"`
	ExcludePatterns []string `json:"exclude_patterns,omitempty"`
	Repositories    []string `json:"repositories,omitempty"`
	Operations      []string `json:"operations,omitempty"`
}

type PermissionTarget struct {
	Name  string      `json:"name"`
	Repo  *Permission `json:"repo,omitempty"`
	Build *Permission `json:"build,omitempty"`
	// ReleaseBundle Permission `json:"release_bundle,omitempty"`
}
