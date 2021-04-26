package artifactorysecrets

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

func TestValidatePermissionTarget(t *testing.T) {

	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		perm := &Permission{
			Repositories: []string{"repo"},
			Operations:   []string{"read", "write", "annotate"},
		}

		pt := PermissionTarget{
			Repo: perm,
		}

		if err := pt.assertValid(); err != nil {
			t.Fatalf("not expecting error. err: %s", err.Error())
		}
	})

	t.Run("empty_repo_repositories", func(t *testing.T) {
		t.Parallel()
		perm := &Permission{
			Operations: []string{"read", "write", "annotate"},
		}
		pt := PermissionTarget{
			Repo: perm,
		}
		err := pt.assertValid()
		if err == nil {
			t.Fatalf("expected error")
		}
		if exp, act := "'repo.repositories' field must be supplied", err.Error(); !strings.Contains(act, exp) {
			t.Errorf("expected %q to match %q", act, exp)
		}
	})

	t.Run("empty_repo_opeartions", func(t *testing.T) {
		t.Parallel()
		perm := &Permission{
			Repositories: []string{"repo"},
		}
		pt := PermissionTarget{
			Repo: perm,
		}
		err := pt.assertValid()
		if err == nil {
			t.Fatalf("expected error")
		}
		if exp, act := "'repo.operations' field must be supplied", err.Error(); !strings.Contains(act, exp) {
			t.Errorf("expected %q to match %q", act, exp)
		}
	})
}

func TestValidateOperations(t *testing.T) {
	t.Parallel()

	t.Run("valid_operations", func(t *testing.T) {
		t.Parallel()

		validOps := []string{"read", "write", "annotate", "delete", "manage", "managedXrayMeta", "distribute"}
		err := validateOperations(validOps)

		if err != nil {
			t.Fatalf("not expecting error. err: %s", err.Error())
		}
	})

	t.Run("invalid_operations", func(t *testing.T) {
		t.Parallel()

		invalidOps := []string{"hello", "world", "read"}
		err := validateOperations(invalidOps)

		if err == nil {
			t.Fatalf("expecting error")
		}

		if merr, ok := err.(*multierror.Error); ok {
			if len(merr.Errors) != 2 {
				t.Errorf("expecting %d errors, got %d", 2, len(merr.Errors))
			}
		}

		if exp := "operation 'hello' is not allowed"; !strings.Contains(err.Error(), exp) {
			t.Errorf("err: %s, exp: %s", err.Error(), exp)
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
			Repo: &Permission{
				Repositories: []string{"repo"},
				Operations:   []string{"read", "write"},
			},
		}
		cpt := &services.PermissionTargetParams{}
		convertPermissionTarget(pt, cpt, groupName(role), "testname")

		if len(cpt.Repo.Actions.Groups) != 1 {
			t.Fatalf("incorrect number of groups")
		}

		if len(cpt.Repo.Actions.Groups["vault-plugin.1234567890"]) != 2 {
			t.Fatalf("incorrect number of operations")
		}

		if got, exp := cpt.Repo.Actions.Groups["vault-plugin.1234567890"], []string{"read", "write"}; !reflect.DeepEqual(got, exp) {
			t.Fatalf("operations don't match: exp: %v, got: %v", exp, got)
		}

	})
}

func envOrDefault(key, d string) string {
	env := os.Getenv(key)
	if env == "" {
		return d
	}
	return env
}
