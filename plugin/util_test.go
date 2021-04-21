package artifactorysecrets

import (
	"reflect"
	"strings"
	"testing"

	v2 "github.com/atlassian/go-artifactory/v2/artifactory/v2"
)

func TestValidatePermissionTarget(t *testing.T) {
	repo := &Permission{
		Repositories: []string{repo},
	}

	t.Parallel()

	t.Run("nil_name", func(t *testing.T) {
		t.Parallel()
		nilPtName := PermissionTarget{
			Repo: repo,
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
		emptyPtName := PermissionTarget{
			Name: emptyString,
			Repo: repo,
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

func TestConvertPermissionTarget(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		role := &RoleStorageEntry{
			Name:   "groupname",
			RoleID: "1234567890",
		}

		pt := &PermissionTarget{
			Name: "testname",
			Repo: &Permission{
				Repositories: []string{"repo"},
				Operations:   []string{"read", "write"},
			},
		}
		cpt := &v2.PermissionTarget{}
		convertPermissionTarget(pt, cpt, role)

		if len(*cpt.Repo.Actions.Groups) != 1 {
			t.Fatalf("incorrect number of groups")
		}

		if len((*cpt.Repo.Actions.Groups)["vault-plugin.1234567890"]) != 2 {
			t.Fatalf("incorrect number of operations")
		}

		if got, exp := (*cpt.Repo.Actions.Groups)["vault-plugin.1234567890"], []string{"read", "write"}; !reflect.DeepEqual(got, exp) {
			t.Fatalf("operations don't match: exp: %v, got: %v", exp, got)
		}

	})
}
