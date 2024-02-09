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
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArtAccPathRole(t *testing.T) {
	t.Parallel()
	if os.Getenv(envVarRunArtAccTests) == "" {
		t.Skip("skipping Artifactory acceptance test (ARTIFACTORY_ACC env var not set)")
	}

	repo := envOrDefault("ARTIFACTORY_REPOSITORY_NAME", "ANY")
	rawPt := fmt.Sprintf(`
	[
		{
			"repo": {
				"include_patterns": ["/mytest/**"],
				"exclude_patterns": [""],
				"repositories": ["%s"],
				"operations": ["read", "write", "annotate"]
			}
		}
	]
	`, repo)

	t.Run("create_role", func(t *testing.T) {
		req, backend := newArtAccEnv(t)

		roleName := "test_create_role"
		data := map[string]interface{}{
			"permission_targets": rawPt,
			"name":               roleName,
		}
		mustRoleCreate(req, backend, t, roleName, data)
	})

	t.Run("create_multiple_role", func(t *testing.T) {
		req, backend := newArtAccEnv(t)

		for i := 1; i < 10; i++ {
			roleName := fmt.Sprintf("role_%d", i)
			data := map[string]interface{}{
				"permission_targets": rawPt,
				"name":               roleName,
			}
			t.Run(roleName, func(t *testing.T) {
				t.Parallel()
				mustRoleCreate(req, backend, t, roleName, data)
			})
		}
	})

	t.Run("get_role", func(t *testing.T) {
		req, backend := newArtAccEnv(t)

		roleName := "test_get_role"
		data := map[string]interface{}{
			"permission_targets": rawPt,
			"name":               roleName,
		}
		mustRoleCreate(req, backend, t, roleName, data)

		resp, err := testRoleRead(req, backend, t, roleName)
		require.NoError(t, err)
		require.False(t, resp.IsError())

		var returnedRole RoleStorageEntry
		err = mapstructure.Decode(resp.Data, &returnedRole)
		require.NoError(t, err, "failed to decode")
		assert.Equal(t, roleName, returnedRole.Name, "incorrect role name %q returned", returnedRole.Name)
	})

	t.Run("update_role_without_change", func(t *testing.T) {
		req, backend := newArtAccEnv(t)
		roleName := "test_update_role"
		data := map[string]interface{}{
			"permission_targets": rawPt,
			"name":               roleName,
		}
		mustRoleCreate(req, backend, t, roleName, data)
		mustRoleCreate(req, backend, t, roleName, data)
	})

	t.Run("list_roles", func(t *testing.T) {
		req, backend := newArtAccEnv(t)

		roleName1 := "test_list_role1"
		roleName2 := "test_list_role2"
		data := map[string]interface{}{
			"permission_targets": rawPt,
			"name":               roleName1,
		}

		mustRoleCreate(req, backend, t, roleName1, data)
		data["name"] = roleName2
		mustRoleCreate(req, backend, t, roleName2, data)

		resp, err := testRoleList(req, backend, t)
		require.NoError(t, err)
		require.False(t, resp.IsError())

		var listResp map[string]interface{}
		err = mapstructure.Decode(resp.Data, &listResp)
		require.NoError(t, err)

		returnedRoles := listResp["keys"].([]string)
		assert.Len(t, returnedRoles, 2, "incorrect number of roles")
		assert.Equal(t, roleName1, returnedRoles[0], "incorrect path set")
		assert.Equal(t, roleName2, returnedRoles[1], "incorrect path set")
	})

	t.Run("delete_role", func(t *testing.T) {
		req, backend := newArtAccEnv(t)

		roleName := "test_delete_role"
		data := map[string]interface{}{
			"permission_targets": rawPt,
			"name":               roleName,
		}

		mustRoleCreate(req, backend, t, roleName, data)
		mustRoleDelete(req, backend, t, roleName)
	})

	t.Run("create role failed with nonexisting repository", func(t *testing.T) {
		req, backend := newArtAccEnv(t)
		roleName := "test_role_nonexisting_repository"
		nonexistingRepoName := "nonexisting_repo"
		rawPt := fmt.Sprintf(`
		[
			{
				"repo": {
					"include_patterns": ["/mytest/**"],
					"exclude_patterns": [""],
					"repositories": ["%s"],
					"operations": ["read"]
				}
			}
		]
		`, nonexistingRepoName)
		data := map[string]interface{}{
			"permission_targets": rawPt,
			"name":               roleName,
		}
		resp, err := testRoleCreate(req, backend, t, roleName, data)
		require.NoError(t, err)
		actualErr := resp.Data["error"].(string)
		expected := fmt.Sprintf("Permission target contains a reference to a non-existing repository '%s'", nonexistingRepoName)
		assert.Contains(t, actualErr, expected)
	})
}

func TestArtAccPermissionTargets(t *testing.T) {
	t.Parallel()
	if os.Getenv(envVarRunArtAccTests) == "" {
		t.Skip("skipping Artifactory acceptance test (ARTIFACTORY_ACC env var not set)")
	}

	ctx := context.Background()
	repo := envOrDefault("ARTIFACTORY_REPOSITORY_NAME", "ANY")
	rawPt := fmt.Sprintf(`
	[
		{
			"repo": {
				"include_patterns": ["/mytest/**"],
				"exclude_patterns": [""],
				"repositories": ["%s"],
				"operations": ["read", "write", "annotate"]
			}
		}
	]
	`, repo)

	t.Run("modify_permission_target", func(t *testing.T) {
		req, backend := newArtAccEnv(t)
		ac := mustGetAccClient(ctx, t, req, backend)

		roleName := "test_modify_permission_target_role"
		data := map[string]interface{}{
			"permission_targets": rawPt,
			"name":               roleName,
		}
		mustRoleCreate(req, backend, t, roleName, data)
		role, err := getRoleEntry(ctx, req.Storage, roleName)
		require.NoError(t, err)

		assertPermissionTarget(t, ac, role, 0)

		modifiedPt := fmt.Sprintf(`
		[
			{
				"repo": {
					"include_patterns": ["/mytest-modified/**"],
					"exclude_patterns": [""],
					"repositories": ["%s"],
					"operations": ["write"]
				}
			}
		]
		`, repo)

		data["permission_targets"] = modifiedPt
		mustRoleUpdate(req, backend, t, roleName, data)
		role, err = getRoleEntry(ctx, req.Storage, roleName)
		require.NoError(t, err)

		// assert permission target in Artifactory matches updated role data
		assertPermissionTarget(t, ac, role, 0)
	})

	t.Run("append_permission_target", func(t *testing.T) {
		req, backend := newArtAccEnv(t)
		ac := mustGetAccClient(ctx, t, req, backend)

		roleName := "test_modify_permission_target_role"
		data := map[string]interface{}{
			"permission_targets": rawPt,
			"name":               roleName,
		}
		mustRoleCreate(req, backend, t, roleName, data)
		role, err := getRoleEntry(ctx, req.Storage, roleName)
		require.NoError(t, err)

		assertPermissionTarget(t, ac, role, 0)

		newPt := fmt.Sprintf(`
		[
			{
				"repo": {
					"include_patterns": ["/mytest/**"],
					"exclude_patterns": [""],
					"repositories": ["%s"],
					"operations": ["read", "write", "annotate"]
				}
			},
			{
				"repo": {
					"include_patterns": ["/mytest2/**"],
					"exclude_patterns": ["/mytest2/foo/**"],
					"repositories": ["%s"],
					"operations": ["read", "write"]
				}
			}
		]
		`, repo, repo)

		data["permission_targets"] = newPt
		mustRoleUpdate(req, backend, t, roleName, data)
		role, err = getRoleEntry(ctx, req.Storage, roleName)
		require.NoError(t, err)

		// assert permission target in Artifactory matches role data
		assertPermissionTarget(t, ac, role, 0)
		assertPermissionTarget(t, ac, role, 1)
	})

	t.Run("delete_permission_target", func(t *testing.T) {
		req, backend := newArtAccEnv(t)
		ac := mustGetAccClient(ctx, t, req, backend)

		roleName := "test_delete_permission_target_role"

		initialPts := fmt.Sprintf(`
		[
			{
				"repo": {
					"include_patterns": ["/mytest/**"],
					"exclude_patterns": [""],
					"repositories": ["%s"],
					"operations": ["read", "write", "annotate"]
				}
			},
			{
				"repo": {
					"include_patterns": ["/mytest2/**"],
					"exclude_patterns": ["/mytest2/foo/**"],
					"repositories": ["%s"],
					"operations": ["read", "write"]
				}
			}
		]
		`, repo, repo)
		data := map[string]interface{}{
			"permission_targets": initialPts,
			"name":               roleName,
		}
		mustRoleCreate(req, backend, t, roleName, data)
		role, err := getRoleEntry(ctx, req.Storage, roleName)
		require.NoError(t, err)

		assertPermissionTarget(t, ac, role, 0)
		assertPermissionTarget(t, ac, role, 1)

		removedPt := fmt.Sprintf(`
		[
			{
				"repo": {
					"include_patterns": ["/mytest/**"],
					"exclude_patterns": [""],
					"repositories": ["%s"],
					"operations": ["read", "write", "annotate"]
				}
			}
		]
		`, repo)

		data["permission_targets"] = removedPt
		mustRoleUpdate(req, backend, t, roleName, data)
		role, err = getRoleEntry(ctx, req.Storage, roleName)
		require.NoError(t, err)
		assert.Len(t, role.PermissionTargets, 1)

		// assert permission target in Artifactory matches role data
		assertPermissionTarget(t, ac, role, 0)
		assertPermissionTargetDeleted(t, ac, role, 1)
	})

	t.Run("delete_role_removes_resources", func(t *testing.T) {
		req, backend := newArtAccEnv(t)
		ac := mustGetAccClient(ctx, t, req, backend)

		roleName := "test_delete_role"

		data := map[string]interface{}{
			"permission_targets": rawPt,
			"name":               roleName,
		}
		mustRoleCreate(req, backend, t, roleName, data)
		role, err := getRoleEntry(ctx, req.Storage, roleName)
		require.NoError(t, err)

		assertPermissionTarget(t, ac, role, 0)

		mustRoleDelete(req, backend, t, roleName)

		assertGroupDeleted(t, ac, role)
		assertPermissionTargetDeleted(t, ac, role, 0)

		role, err = getRoleEntry(ctx, req.Storage, roleName)
		require.Nil(t, role)
		require.NoError(t, err)
	})
}

func TestPathRoleFail(t *testing.T) {
	t.Parallel()
	req, backend := newArtMockEnv(t)
	conf := map[string]interface{}{
		"base_url":     "https://example.jfrog.io/example",
		"bearer_token": "mybearertoken",
		"max_ttl":      "3600s",
	}
	testConfigUpdate(t, backend, req.Storage, conf)

	t.Run("nonexistent_role", func(t *testing.T) {
		data := make(map[string]interface{})
		roleName := "test_role1"
		data["name"] = roleName
		resp, err := testRoleRead(req, backend, t, "noname")
		require.NoError(t, err)
		require.Nil(t, resp)
	})

	t.Run("exceed_config_max_ttl", func(t *testing.T) {
		roleName := "test_role_max_ttl"
		data := make(map[string]interface{})
		rawPt := `
		[
			{
				"repo": {
					"include_patterns": ["/mytest/**"],
					"exclude_patterns": [""],
					"repositories": ["ANY"],
					"operations": ["read", "write", "annotate"]
				}
			}
		]
		`
		data["name"] = roleName
		data["max_ttl"] = "7200s"
		data["permission_targets"] = rawPt
		resp, err := testRoleCreate(req, backend, t, roleName, data)
		require.NoError(t, err)
		require.True(t, resp.IsError(), "expecting error")

		actualErr := resp.Data["error"].(string)
		expected := "role max ttl is greater than config max ttl"
		assert.Contains(t, actualErr, expected)
	})

	t.Run("exceed_role_max_ttl", func(t *testing.T) {
		data := make(map[string]interface{})
		roleName := "test_role_token_ttl"
		rawPt := `
		[
			{
				"repo": {
					"include_patterns": ["/mytest/**"],
					"exclude_patterns": [""],
					"repositories": ["ANY"],
					"operations": ["read", "write", "annotate"]
				}
			}
		]
		`
		data["name"] = roleName
		data["token_ttl"] = "7200s"
		data["permission_targets"] = rawPt
		resp, err := testRoleCreate(req, backend, t, roleName, data)
		require.NoError(t, err)
		require.True(t, resp.IsError(), "expecting error")

		actualErr := resp.Data["error"].(string)
		expected := "role token ttl is greater than role max ttl"
		assert.Contains(t, actualErr, expected)
	})

	t.Run("no_permission_targets_for_new_role", func(t *testing.T) {
		data := make(map[string]interface{})
		roleName := "test_role1"
		resp, err := testRoleCreate(req, backend, t, roleName, data)
		require.NoError(t, err)
		require.True(t, resp.IsError(), "expecting error")

		actualErr := resp.Data["error"].(string)
		expected := "permission targets are required for new role"
		assert.Contains(t, actualErr, expected)
	})

	t.Run("empty_permission_targets", func(t *testing.T) {
		data := make(map[string]interface{})
		roleName := "test_role1"
		data["permission_targets"] = ""
		resp, err := testRoleCreate(req, backend, t, roleName, data)
		require.NoError(t, err)
		require.True(t, resp.IsError(), "expecting error")

		actualErr := resp.Data["error"].(string)
		expected := "permission targets are empty"
		assert.Contains(t, actualErr, expected)
	})

	t.Run("unmarshable_permission_target", func(t *testing.T) {
		data := make(map[string]interface{})
		roleName := "test_role1"
		data["permission_targets"] = 60
		data["name"] = roleName
		resp, err := testRoleCreate(req, backend, t, roleName, data)
		require.NoError(t, err)
		require.True(t, resp.IsError(), "expecting error")

		actualErr := resp.Data["error"].(string)
		expected := "Error unmarshal permission targets. Expecting list of permission targets"
		assert.Contains(t, actualErr, expected)
	})

	t.Run("permission_target_empty_required_field", func(t *testing.T) {
		data := make(map[string]interface{})
		roleName := "test_role1"
		rawPt := `
		[
			{
				"repo": {
					"include_patterns": ["/mytest/**"],
					"exclude_patterns": [""]
				}
			}
		]
		`
		data["permission_targets"] = rawPt
		data["name"] = roleName
		resp, err := testRoleCreate(req, backend, t, roleName, data)
		require.NoError(t, err)
		require.True(t, resp.IsError(), "expecting error")

		actualErr := resp.Data["error"].(string)
		expected1 := "repo.repositories' field must be supplied"
		expected2 := "repo.operations' field must be supplied"
		assert.Contains(t, actualErr, expected1)
		assert.Contains(t, actualErr, expected2)
	})

	t.Run("permission_target_invalid_operation", func(t *testing.T) {
		data := make(map[string]interface{})
		roleName := "test_role1"
		rawPt := `
		[
			{
				"repo": {
					"include_patterns": ["/mytest/**"],
					"exclude_patterns": [""],
					"operations": ["invalidop"]
				}
			}
		]
		`
		data["permission_targets"] = rawPt
		data["name"] = roleName
		resp, err := testRoleCreate(req, backend, t, roleName, data)
		require.NoError(t, err)
		require.True(t, resp.IsError(), "expecting error")
		actualErr := resp.Data["error"].(string)
		expected := "operation 'invalidop' is not allowed"
		assert.Contains(t, actualErr, expected)
	})
}

// assertPermissionTarget inspects the actual PermissionTarget in Artifactory against the one in vault role.
func assertPermissionTarget(t *testing.T, ac artifactory.ArtifactoryServicesManager, role *RoleStorageEntry, permissionTargetIndex int) {
	t.Helper()
	ptName := permissionTargetName(role.Name, permissionTargetIndex)
	expected := role.PermissionTargets[permissionTargetIndex]
	actual, err := ac.GetPermissionTarget(ptName)
	require.NoError(t, err, "Error retrieving permission target from Artifactory")

	assert.Equal(t, expected.Repo.IncludePatterns, actual.Repo.IncludePatterns, "permission target IncludePatterns should match permission target input provided to vault")
	if len(expected.Repo.ExcludePatterns) == 1 && expected.Repo.ExcludePatterns[0] == "" {
		assert.Empty(t, actual.Repo.ExcludePatterns, "permission target ExcludePatterns should be empty")
	} else {
		assert.Equal(t, expected.Repo.ExcludePatterns, actual.Repo.ExcludePatterns, "permission target ExcludePatterns should match permission target input provided to vault")

	}
	assert.Equal(t, expected.Repo.Repositories, actual.Repo.Repositories, "permission target repositories should match permission target input provided to vault")

	assert.Len(t, actual.Repo.Actions.Groups, 1, "A generated Permission Target should map to a single group.")
	actualGroupOperations := actual.Repo.Actions.Groups[groupName(role)]
	assert.ElementsMatch(t, expected.Repo.Operations, actualGroupOperations)
}

func assertPermissionTargetDeleted(t *testing.T, ac artifactory.ArtifactoryServicesManager, role *RoleStorageEntry, permissionTargetIndex int) {
	t.Helper()
	ptName := permissionTargetName(role.Name, permissionTargetIndex)
	actual, err := ac.GetPermissionTarget(ptName)
	assert.Nil(t, actual)
	assert.NoError(t, err)
}

func assertGroupDeleted(t *testing.T, ac artifactory.ArtifactoryServicesManager, role *RoleStorageEntry) {
	t.Helper()
	params := services.GroupParams{
		GroupDetails: services.Group{
			Name: groupName(role),
		},
	}
	group, err := ac.GetGroup(params)
	assert.NoError(t, err, "Artifactory should return nil error for non-existent group")
	assert.Nil(t, group, "Group %s should be deleted", groupName(role))
}

func testRoleCreate(req *logical.Request, b logical.Backend, t *testing.T, roleName string, data map[string]interface{}) (*logical.Response, error) {
	t.Helper()
	req.Operation = logical.CreateOperation
	req.Path = fmt.Sprintf("roles/%s", roleName)
	req.Data = data

	resp, err := b.HandleRequest(context.Background(), req)
	return resp, err
}

func mustRoleCreate(req *logical.Request, b logical.Backend, t *testing.T, roleName string, data map[string]interface{}) {
	t.Helper()
	resp, err := testRoleCreate(req, b, t, roleName, data)
	require.NoError(t, err)
	require.False(t, resp.IsError())
}

// testRoleUpdate should effectively be the same op as testRoleCreate as the same
// pathRoleCreateUpdate function is used for both Create and Update operations.
func testRoleUpdate(req *logical.Request, b logical.Backend, t *testing.T, roleName string, data map[string]interface{}) (*logical.Response, error) {
	t.Helper()
	req.Operation = logical.UpdateOperation
	req.Path = fmt.Sprintf("roles/%s", roleName)
	req.Data = data

	resp, err := b.HandleRequest(context.Background(), req)
	return resp, err
}

func mustRoleUpdate(req *logical.Request, b logical.Backend, t *testing.T, roleName string, data map[string]interface{}) {
	t.Helper()
	resp, err := testRoleUpdate(req, b, t, roleName, data)
	require.NoError(t, err)
	require.False(t, resp.IsError())
}

func testRoleRead(req *logical.Request, b logical.Backend, t *testing.T, roleName string) (*logical.Response, error) {
	t.Helper()
	data := map[string]interface{}{
		"name": roleName,
	}

	req.Operation = logical.ReadOperation
	req.Path = fmt.Sprintf("roles/%s", roleName)
	req.Data = data

	resp, err := b.HandleRequest(context.Background(), req)
	return resp, err
}

func testRoleDelete(req *logical.Request, b logical.Backend, t *testing.T, roleName string) (*logical.Response, error) {
	t.Helper()
	data := map[string]interface{}{
		"name": roleName,
	}

	req.Operation = logical.DeleteOperation
	req.Path = fmt.Sprintf("roles/%s", roleName)
	req.Data = data

	resp, err := b.HandleRequest(context.Background(), req)
	return resp, err
}

func mustRoleDelete(req *logical.Request, b logical.Backend, t *testing.T, roleName string) {
	resp, err := testRoleDelete(req, b, t, roleName)
	require.NoError(t, err)
	require.Nil(t, resp)
}

func testRoleList(req *logical.Request, b logical.Backend, t *testing.T) (*logical.Response, error) {
	t.Helper()
	req.Operation = logical.ListOperation
	req.Path = "roles"
	resp, err := b.HandleRequest(context.Background(), req)
	return resp, err
}
