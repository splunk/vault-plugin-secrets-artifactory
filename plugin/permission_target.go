package artifactorysecrets

import (
	"errors"

	"github.com/hashicorp/go-multierror"
)

type Permission struct {
	IncludePatterns []string `json:"include_patterns,omitempty"`
	ExcludePatterns []string `json:"exclude_patterns,omitempty"`
	Repositories    []string `json:"repositories,omitempty"`
	Operations      []string `json:"operations,omitempty"`
}

type PermissionTarget struct {
	Repo  *Permission `json:"repo,omitempty"`
	Build *Permission `json:"build,omitempty"`
	// ReleaseBundle Permission `json:"release_bundle,omitempty"`
}

// validate user supplied permission target
func (pt PermissionTarget) assertValid() error {
	var err *multierror.Error

	if pt.Repo != nil {
		if len(pt.Repo.Repositories) == 0 {
			err = multierror.Append(err, errors.New("'repo.repositories' field must be supplied"))
		}
		if len(pt.Repo.Operations) == 0 {
			err = multierror.Append(err, errors.New("'repo.operations' field must be supplied"))
		} else if e := validateOperations(pt.Repo.Operations); e != nil {
			err = multierror.Append(err, e)
		}
	}

	if pt.Build != nil {
		if len(pt.Build.Repositories) == 0 {
			err = multierror.Append(err, errors.New("'build.repositories' field must be supplied"))
		}
		if len(pt.Build.Operations) == 0 {
			err = multierror.Append(err, errors.New("'build.operations' field must be supplied"))
		} else if e := validateOperations(pt.Build.Operations); e != nil {
			err = multierror.Append(err, e)
		}
	}
	return err.ErrorOrNil()
}
