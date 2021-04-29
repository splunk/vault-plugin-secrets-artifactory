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

	resp, err := testIssueToken(req, backend, t, roleName)
	require.NoError(t, err)
	require.False(t, resp.IsError())

	assert.NotEmpty(t, resp.Data["access_token"], "no token returned")
	assert.NotEmpty(t, resp.Data["username"], "no username returned")
}

// create the token given the parameters
func testIssueToken(req *logical.Request, b logical.Backend, t *testing.T, roleName string) (*logical.Response, error) {
	req.Operation = logical.UpdateOperation
	req.Path = fmt.Sprintf("token/%s", roleName)

	resp, err := b.HandleRequest(context.Background(), req)

	return resp, err
}
