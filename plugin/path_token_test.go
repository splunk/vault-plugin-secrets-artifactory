package artifactorysecrets

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccIssueToken(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test (short)")
	}

	req, backend := newAccEnv(t)

	repo := envOrDefault("ARTIFACTORY_REPOSITORY_NAME", "ANY")

	roleName := "test_issue_token_role"
	data := map[string]interface{}{
		"name": roleName,
		"permission_targets": fmt.Sprintf(`
		[
			{
				"repo": {
					"include_patterns": ["/mytest/**"] ,
					"exclude_patterns": [""],
					"repositories": ["%s"],
					"operations": ["read", "write", "annotate"]
				}
			}
		]
		`, repo),
	}
	mustRoleCreate(req, backend, t, roleName, data)

	t.Run("success", func(t *testing.T) {
		d := make(map[string]interface{})
		d["role_name"] = roleName
		resp, err := testIssueToken(req, backend, t, roleName, d)
		require.NoError(t, err)
		require.False(t, resp.IsError())

		assert.NotEmpty(t, resp.Data["access_token"], "no token returned")
		assert.NotEmpty(t, resp.Data["username"], "no username returned")
	})

	t.Run("exceed_ttl", func(t *testing.T) {
		d := make(map[string]interface{})
		d["role_name"] = roleName
		d["ttl"] = "7200s"
		resp, err := testIssueToken(req, backend, t, roleName, d)
		require.NoError(t, err)
		require.True(t, resp.IsError(), "expecting error")

		actualErr := resp.Data["error"].(string)
		expected := "Token ttl is greater than role max ttl"
		assert.Contains(t, actualErr, expected)
	})

}

// create the token given the parameters
func testIssueToken(req *logical.Request, b logical.Backend, t *testing.T, roleName string, data map[string]interface{}) (*logical.Response, error) {
	req.Operation = logical.UpdateOperation
	req.Path = fmt.Sprintf("token/%s", roleName)
	req.Data = data

	resp, err := b.HandleRequest(context.Background(), req)

	return resp, err
}
