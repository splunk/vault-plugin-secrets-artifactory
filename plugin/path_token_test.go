package artifactorysecrets

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
)

func TestIssueToken(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping intergartion test (short)")
	}

	b, storage := getTestBackend(t)

	conf := map[string]interface{}{
		"base_url":     os.Getenv("ARTIFACTORY_URL"),
		"bearer_token": os.Getenv("BEARER_TOKEN"),
		"max_ttl":      "600s",
	}
	testConfigUpdate(t, b, storage, conf)

	req := &logical.Request{
		Storage: storage,
	}
	roleName := "test_role"
	pt := fmt.Sprintf(`
	[
		{
			"name": "test",
			"repo": {
				"include_patterns": ["/mytest/**"] ,
				"exclude_patterns": [""],
				"repositories": ["%s"],
				"operations": ["read", "write", "annotate"]
			}
		}
	]
	`, repo)
	resp, err := testRoleCreate(req, b, t, roleName, pt)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%s resp:%#v\n", err, resp)
	}

	resp, err = testIssueToken(req, b, t, roleName)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%s resp:%#v\n", err, resp)
	}

	if resp.Data["access_token"] == "" {
		t.Fatal("no token returned\n")
	}

	if resp.Data["username"] == "" {
		t.Fatal("no username returned\n")
	}

}

// create the token given the parameters
func testIssueToken(req *logical.Request, b logical.Backend, t *testing.T, roleName string) (*logical.Response, error) {
	req.Operation = logical.UpdateOperation
	req.Path = fmt.Sprintf("token/%s", roleName)

	resp, err := b.HandleRequest(context.Background(), req)

	return resp, err
}
