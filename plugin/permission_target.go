// Copyright  2021 Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
