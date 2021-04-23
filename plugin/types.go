package artifactorysecrets

type Permission struct {
	IncludePatterns []string `json:"include_patterns,omitempty"`
	ExcludePatterns []string `json:"exclude_patterns,omitempty"`
	Repositories    []string `json:"repositories"`
	Operations      []string `json:"operations"`
}

type PermissionTarget struct {
	Repo  *Permission `json:"repo,omitempty"`
	Build *Permission `json:"build,omitempty"`
	// ReleaseBundle Permission `json:"release_bundle,omitempty"`
}
