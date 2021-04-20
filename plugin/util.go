package artifactorysecrets

import (
	"fmt"

	v2 "github.com/atlassian/go-artifactory/v2/artifactory/v2"
)

const (
	groupNamePrefix     = "vault-plugin"
	tokenUsernamePrefix = "auto-vault-plugin"
	pluginOwnRole       = "VAULT_PLUGIN_OWN_ROLE"
)

func groupName(roleName string) string {
	return fmt.Sprintf("%s.%s", groupNamePrefix, roleName)
}

func permissionTargetName(roleEntry *RoleStorageEntry, name string) *string {
	n := fmt.Sprintf("%s.%s", name, roleEntry.RoleID)
	return &n
}
func tokenUsername(role_name string) string {
	return fmt.Sprintf("%s.%s", tokenUsernamePrefix, role_name)
}

// replaceGroupName swaps pluginOwnRole with supplied group name
// this is crashing
func replaceGroupName(pt *v2.PermissionTarget, groupName string) {
	if pt.Repo != nil && pt.Repo.Actions != nil && pt.Repo.Actions.Groups != nil {
		for name, Ops := range *pt.Repo.Actions.Groups {
			if name == pluginOwnRole {
				delete(*pt.Repo.Actions.Groups, name)
				(*pt.Repo.Actions.Groups)[groupName] = Ops
				break
			}
		}
	}

	if pt.Build != nil && pt.Build.Actions != nil && pt.Build.Actions.Groups != nil {
		for name, Ops := range *pt.Build.Actions.Groups {
			if name == pluginOwnRole {
				delete(*pt.Repo.Actions.Groups, name)
				(*pt.Build.Actions.Groups)[groupName] = Ops
			}
		}
	}

}
