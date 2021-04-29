package artifactorysecrets

import (
	"os"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

		err := pt.assertValid()
		require.NoError(t, err, "not expecting error: %s", err)
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
		require.Error(t, err, "expecting error")
		assert.Contains(t, err.Error(), "'repo.repositories' field must be supplied")
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
		require.Error(t, err, "expecting error")
		assert.Contains(t, err.Error(), "'repo.operations' field must be supplied")
	})
}

func TestValidateOperations(t *testing.T) {
	t.Parallel()

	t.Run("valid_operations", func(t *testing.T) {
		t.Parallel()

		validOps := []string{"read", "write", "annotate", "delete", "manage", "managedXrayMeta", "distribute"}
		err := validateOperations(validOps)

		require.NoError(t, err, "not expecting error: %s", err)
	})

	t.Run("invalid_operations", func(t *testing.T) {
		t.Parallel()

		invalidOps := []string{"hello", "world", "read"}
		err := validateOperations(invalidOps)
		require.Error(t, err, "expecting error")

		if merr, ok := err.(*multierror.Error); ok {
			assert.Len(t, merr.Errors, 2, "expecting %d errors, got %d", 2, len(merr.Errors))
		}

		assert.Contains(t, err.Error(), "operation 'hello' is not allowed")
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

		assert.Len(t, cpt.Repo.Actions.Groups, 1, "incorrect number of groups")
		assert.Len(t, cpt.Repo.Actions.Groups["vault-plugin.1234567890"], 2, "incorrect number of operations")
		assert.ElementsMatch(t, []string{"read", "write"}, cpt.Repo.Actions.Groups["vault-plugin.1234567890"])
	})
}

func envOrDefault(key, d string) string {
	env := os.Getenv(key)
	if env == "" {
		return d
	}
	return env
}
