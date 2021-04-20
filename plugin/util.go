package artifactorysecrets

import (
	"fmt"
)

const (
	groupNamePrefix = "vault-plugin"
	pluginOwnRole   = "VAULT_PLUGIN_OWN_ROLE"
)

func groupName(roleName string) string {
	return fmt.Sprintf("%s.%s", groupNamePrefix, roleName)
}

func permissionTargetName(roleEntry *RoleStorageEntry, name string) *string {
	n := fmt.Sprintf("%s.%s", name, roleEntry.RoleID)
	return &n
}

// replaceGroupName swaps pluginOwnRole with supplied group name
// this is crashing
// func replaceGroupName(pt *v2.PermissionTarget, groupName string) *v2.PermissionTarget {
// 	modifiedPt := pt

// 	repoGroups := *pt.Repo.Actions.Groups
// 	for groupName, Ops := range repoGroups {
// 		if groupName == pluginOwnRole {
// 			delete(*pt.Repo.Actions.Groups, groupName)
// 			repoGroups[groupName] = Ops
// 		}
// 	}

// 	modifiedPt.Repo.Actions.Groups = &repoGroups

// 	buildGroups := *pt.Build.Actions.Groups
// 	for groupName, Ops := range buildGroups {
// 		if groupName == pluginOwnRole {
// 			delete(*pt.Repo.Actions.Groups, groupName)
// 			buildGroups[groupName] = Ops
// 		}
// 	}

// 	modifiedPt.Build.Actions.Groups = &buildGroups

// 	return modifiedPt
// }
