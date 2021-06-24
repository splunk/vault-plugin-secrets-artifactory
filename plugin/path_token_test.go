// Copyright  2021 Splunk, Inc.
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArtAccIssueToken(t *testing.T) {
	if os.Getenv(envVarRunArtAccTests) == "" {
		t.Skip("skipping Artifactory acceptance test (ARTIFACTORY_ACC env var not set)")
	}

	req, backend := newArtAccEnv(t)

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
