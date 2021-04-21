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

func permissionTargetName(roleEntry *RoleStorageEntry, ptName string) *string {
	n := fmt.Sprintf("%s.%s.%s", pluginPrefix, ptName, roleEntry.RoleID)
	return &n
}
func tokenUsername(roleName string) string {
	return fmt.Sprintf("%s.%s", tokenUsernamePrefix, roleName)
}

// validatePermissionTarget checks on necessary fields in permission target
func validatePermissionTarget(pt *PermissionTarget) error {

	if pt.Name == "" {
		return fmt.Errorf("'name' field must be supplied")
	}

	return nil
}

func convertPermissionTarget(fromPt *PermissionTarget, toPt *v2.PermissionTarget, roleEntry *RoleStorageEntry) {

	if fromPt.Repo != nil {
		groupRepo := map[string][]string{
			groupName(roleEntry): fromPt.Repo.Operations,
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
			groupName(roleEntry): fromPt.Build.Operations,
		}
		p := &v2.Permission{
			IncludePatterns: &fromPt.Build.IncludePatterns,
			ExcludePatterns: &fromPt.Build.ExcludePatterns,
			Repositories:    &fromPt.Build.Repositories,
			Actions:         &v2.Entity{Groups: &groupBuild},
		}
		toPt.Build = p
	}

	toPt.Name = &fromPt.Name
}
