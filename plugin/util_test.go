package artifactorysecrets

import (
	"reflect"
	"testing"

	v2 "github.com/atlassian/go-artifactory/v2/artifactory/v2"
)

func TestReplaceGroupName(t *testing.T) {
	groupName := "role1234"
	perms := []string{"read", "annotate", "write"}
	groups := map[string][]string{
		"VAULT_PLUGIN_OWN_ROLE": perms,
	}
	entity := v2.Entity{
		Groups: &groups,
	}
	pt := v2.PermissionTarget{
		Repo: &v2.Permission{
			Actions: &entity,
		},
	}

	replaceGroupName(&pt, groupName)

	if len(*pt.Repo.Actions.Groups) != 1 {
		t.Fatalf("incorrect number of groups")
	}

	if p, ok := (*pt.Repo.Actions.Groups)[groupName]; ok {
		if !reflect.DeepEqual(p, perms) {
			t.Fatalf("permission doesn't match. exp: %v got: %v", perms, p)
		}
	} else {
		t.Fatalf("expected group name %s doesn't exist", groupName)
	}
}
