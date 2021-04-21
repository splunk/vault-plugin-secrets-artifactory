package artifactorysecrets

import (
	"reflect"
	"strings"
	"testing"

	v2 "github.com/atlassian/go-artifactory/v2/artifactory/v2"
)

func TestValidatePermissionTarget(t *testing.T) {
	repo := v2.Permission{
		Repositories: &[]string{repo},
	}

	t.Parallel()

	t.Run("nil_name", func(t *testing.T) {
		t.Parallel()
		nilPtName := v2.PermissionTarget{
			Repo: &repo,
		}
		err := validatePermissionTarget(&nilPtName)
		if err == nil {
			t.Fatalf("expected error")
		}
		if exp, act := "'name' field must be supplied", err.Error(); !strings.EqualFold(act, exp) {
			t.Errorf("expected %q to match %q", act, exp)
		}
	})

	t.Run("empty_name", func(t *testing.T) {
		t.Parallel()

		emptyString := ""
		emptyPtName := v2.PermissionTarget{
			Name: &emptyString,
			Repo: &repo,
		}
		err := validatePermissionTarget(&emptyPtName)
		if err == nil {
			t.Fatalf("expected error")
		}
		if exp, act := "'name' field must be supplied", err.Error(); !strings.EqualFold(act, exp) {
			t.Errorf("expected %q to match %q", act, exp)
		}
	})

}

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
