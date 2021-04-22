package artifactorysecrets

import (
	"fmt"

	v2 "github.com/atlassian/go-artifactory/v2/artifactory/v2"
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

func convertPermissionTarget(fromPt *PermissionTarget, toPt *v2.PermissionTarget, groupName, ptName string) {

	if fromPt.Repo != nil {
		groupRepo := map[string][]string{
			groupName: fromPt.Repo.Operations,
		}
		p := &v2.Permission{
			IncludePatterns: &fromPt.Repo.IncludePatterns,
			ExcludePatterns: &fromPt.Repo.ExcludePatterns,
			Repositories:    &fromPt.Repo.Repositories,
			Actions:         &v2.Entity{Groups: &groupRepo},
		}
		toPt.Repo = p
	}

	if fromPt.Build != nil {

		groupBuild := map[string][]string{
			groupName: fromPt.Build.Operations,
		}
		p := &v2.Permission{
			IncludePatterns: &fromPt.Build.IncludePatterns,
			ExcludePatterns: &fromPt.Build.ExcludePatterns,
			Repositories:    &fromPt.Build.Repositories,
			Actions:         &v2.Entity{Groups: &groupBuild},
		}
		toPt.Build = p
	}

	toPt.Name = &ptName
}

// validate user supplied permission target
func (pt PermissionTarget) assertValid() error {
	// if pt.Name == "" {
	// 	return fmt.Errorf("'name' field must be supplied")
	// }

	return nil
}
