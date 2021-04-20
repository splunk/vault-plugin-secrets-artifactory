package artifactorysecrets

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
)

func TestIssueValidateToken(t *testing.T) {
	b, storage := getTestBackend(t)

	conf := map[string]interface{}{
		"base_url":     "https://example.jfrog.io/example",
		"bearer_token": "mybearertoken",
		"max_ttl":      "600s",
	}
	testConfigUpdate(t, b, storage, conf)

	req := &logical.Request{
		Storage: storage,
	}
	roleName := "test_role"
	resp, err := testRoleCreate(req, b, t, roleName)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%s resp:%#v\n", err, resp)
	}

	/* Enable this when we have a temp artifactory bearer token */
	//resp, err = testTokenIssue(req, b, t, roleName, entryName)
	//if err != nil || (resp != nil && resp.IsError()) {
	//	t.Fatalf("err:%s resp:%#v\n", err, resp)
	//}
	//
	//if resp.Data["access_token"] == "" {
	//	t.Fatal("no token returned\n")
	//}
	//
	//if resp.Data["username"] == "" {
	//	t.Fatal("no username returned\n")
	//}

}

// create the token given the parameters
func testTokenIssue(req *logical.Request, b logical.Backend, t *testing.T, roleName string) (*logical.Response, error) {
	data := map[string]interface{}{
		"role_name": roleName,
		"path":      "testpath1",
	}

	req.Operation = logical.UpdateOperation
	req.Path = fmt.Sprintf("issue/%s", roleName)
	req.Data = data

	resp, err := b.HandleRequest(context.Background(), req)

	return resp, err
}
