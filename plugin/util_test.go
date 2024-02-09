// Copyright  2024 Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package artifactorysecrets

import (
	"os"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const maxArtifactoryNameLen = 64

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

func TestRoleID(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
	}{
		{
			name:   "roleID is less than 64 chars (artifactory max)",
			input:  "rolename-long-but-less-than-max",
			maxLen: maxArtifactoryNameLen,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := roleID(test.input)
			checkTokenUsernameLength(t, actual)
			if len(actual) > test.maxLen {
				t.Errorf("roleID: %v, len(roleID): %v", actual, len(actual))
			}
		})

	}
}

func TestTokenUserName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "token username less than max size",
			input: "rolename-long-but-less-than-max",
			want:  "auto-vault-plugin.rolename-long-but-less-than-max",
		},
		{
			name:  "token username less than max size",
			input: "rolename-too-long-to-fit-into-artifactory-token-username",
			want:  "auto-vault-plugin.rolename-too-long-to-fit-into-ara8c837ff",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := tokenUsername(test.input)
			checkTokenUsernameLength(t, got)
			if got != test.want {
				t.Errorf("token username: %v, want %v", got, test.want)
			}
		})

	}
}

func checkTokenUsernameLength(t *testing.T, username string) {
	if len(username) > tokenUsernameMaxLen {
		t.Errorf("Expected token username to be less than or equal to %v, actual name '%v'", tokenUsernameMaxLen, username)
	}
}

func envOrDefault(key, d string) string {
	env := os.Getenv(key)
	if env == "" {
		return d
	}
	return env
}
