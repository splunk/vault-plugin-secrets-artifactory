package artifactorysecrets

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

const (
	pluginPrefix        = "vault-plugin"
	tokenUsernamePrefix = "auto-vault-plugin"
)

func groupName(roleEntry *RoleStorageEntry) string {
	return fmt.Sprintf("%s.%s", pluginPrefix, roleEntry.RoleID)
}

func permissionTargetName(roleEntry *RoleStorageEntry, index int) string {
	return fmt.Sprintf("%s.pt%d.%s", pluginPrefix, index, roleEntry.RoleID)
}

func tokenUsername(roleName string) string {
	return fmt.Sprintf("%s.%s", tokenUsernamePrefix, roleName)
}

// appendTrailingSlash appends trailing slash if url doesn't end with slash.
// artifactory client assumes URL ends with '/'
func appendTrailingSlash(url string) string {
	if !strings.HasSuffix(url, "/") {
		return fmt.Sprintf("%s/", url)
	}
	return url
}

func convertPermissionTarget(fromPt *PermissionTarget, toPt *services.PermissionTargetParams, groupName, ptName string) {

	if fromPt.Repo != nil {
		groupRepo := map[string][]string{
			groupName: fromPt.Repo.Operations,
		}
		p := &services.PermissionTargetSection{
			IncludePatterns: fromPt.Repo.IncludePatterns,
			ExcludePatterns: fromPt.Repo.ExcludePatterns,
			Repositories:    fromPt.Repo.Repositories,
			Actions:         &services.Actions{Groups: groupRepo},
		}
		toPt.Repo = p
	}

	if fromPt.Build != nil {

		groupBuild := map[string][]string{
			groupName: fromPt.Build.Operations,
		}
		p := &services.PermissionTargetSection{
			IncludePatterns: fromPt.Build.IncludePatterns,
			ExcludePatterns: fromPt.Build.ExcludePatterns,
			Repositories:    fromPt.Build.Repositories,
			Actions:         &services.Actions{Groups: groupBuild},
		}
		toPt.Build = p
	}

	toPt.Name = ptName
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

func validateOperations(ops []string) error {
	var err *multierror.Error

	for _, op := range ops {
		switch op {
		case "read", "write", "annotate",
			"delete", "manage", "managedXrayMeta",
			"distribute":
			continue
		default:
			err = multierror.Append(err, fmt.Errorf("operation '%s' is not allowed", op))
		}
	}

	return err.ErrorOrNil()
}
