package artifactorysecrets

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/mitchellh/mapstructure"
)

func TestPathRole(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test (short)")
	}

	backend, storage := getTestBackend(t)
	conf := map[string]interface{}{
		"base_url":     os.Getenv("ARTIFACTORY_URL"),
		"bearer_token": os.Getenv("ARTIFACTORY_BEARER_TOKEN"),
		"max_ttl":      "600s",
	}

	testConfigUpdate(t, backend, storage, conf)

	/***  TEST CREATE OPERATION ***/
	req := &logical.Request{
		Storage: storage,
	}
	var repo = envOrDefault("ARTIFACTORY_REPOSITORY_NAME", "ANY")

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
	data := map[string]interface{}{
		"permission_targets": rawPt,
	}

	data["name"] = "test_role1"
	resp, err := testRoleCreate(req, backend, t, "test_role1", data)

	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%s resp:%#v\n", err, resp)
	}

	/***  TEST GET OPERATION ***/
	resp, err = testRoleRead(req, backend, t, "test_role1")
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%s resp:%#v\n", err, resp)
	}

	var returnedRole RoleStorageEntry
	err = mapstructure.Decode(resp.Data, &returnedRole)
	if err != nil {
		t.Fatalf("failed to decode. err: %s", err)
	}

	if returnedRole.Name != "test_role1" {
		t.Fatalf("incorrect role name %s returned, not the same as saved value \n", returnedRole.Name)
	}

	/*** Test Update without update ***/
	data["name"] = "test_role2"
	resp, err = testRoleCreate(req, backend, t, "test_role2", data)

	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%s resp:%#v\n", err, resp)
	}

	resp, err = testRoleCreate(req, backend, t, "test_role2", data)

	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%s resp:%#v\n", err, resp)
	}

	/*** TEST GET NON-EXISTENT ROLE ***/
	resp, err = testRoleRead(req, backend, t, "test_role3")
	if err != nil && resp != nil {
		t.Fatalf("err:%s resp:%#v\n", err, resp)
	}

	/***  TEST List OPERATION ***/
	resp, err = testRoleList(req, backend, t)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%s resp:%#v\n", err, resp)
	}

	var listResp map[string]interface{}
	err = mapstructure.Decode(resp.Data, &listResp)
	if err != nil {
		t.Fatalf("failed to decode. err: %s", err)
	}

	returnedRoles := listResp["keys"].([]string)

	if len(returnedRoles) != 2 {
		t.Fatalf("incorrect number of roles \n")
	}

	if returnedRoles[0] != "test_role1" && returnedRoles[1] != "test_role2" {
		t.Fatalf("incorrect path set \n")
	}

	/***  TEST Delete OPERATION ***/
	resp, err = testRoleDelete(req, backend, t, "test_role1")
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%s resp:%#v\n", err, resp)
	}
}

func TestPathRoleFail(t *testing.T) {

	backend, storage := getTestBackend(t)

	conf := map[string]interface{}{
		"base_url":     os.Getenv("ARTIFACTORY_URL"),
		"bearer_token": os.Getenv("ARTIFACTORY_BEARER_TOKEN"),
		"max_ttl":      "600s",
	}

	testConfigUpdate(t, backend, storage, conf)

	req := &logical.Request{
		Storage: storage,
	}
	data := make(map[string]interface{})
	data["name"] = "test_role1"

	t.Run("no_permission_targets_for_new_role", func(t *testing.T) {

		resp, _ := testRoleCreate(req, backend, t, "test_role1", data)

		if !resp.IsError() {
			t.Fatal("expecting error")
		}
		if errmsg, exp := resp.Data["error"].(string), "permission targets are required for new role"; errmsg != exp {
			t.Errorf("err:%s exp:%#v\n", errmsg, exp)
		}
	})

	t.Run("empty_permission_targets", func(t *testing.T) {

		data["permission_targets"] = ""
		resp, _ := testRoleCreate(req, backend, t, "test_role1", data)

		if !resp.IsError() {
			t.Fatal("expecting error")
		}
		if errmsg, exp := resp.Data["error"].(string), "permission targets are empty"; errmsg != exp {
			t.Errorf("err:%s exp:%#v\n", errmsg, exp)
		}
	})

	t.Run("unmarshable_permission_target", func(t *testing.T) {

		data["permission_targets"] = 60
		resp, _ := testRoleCreate(req, backend, t, "test_role1", data)

		if !resp.IsError() {
			t.Fatal("expecting error")
		}
		if errmsg, exp := resp.Data["error"].(string), "Error unmarshal permission targets. Expecting list of permission targets"; !strings.Contains(errmsg, exp) {
			t.Errorf("err:%s exp:%#v\n", errmsg, exp)
		}
	})

	t.Run("permission_target_empty_required_field", func(t *testing.T) {

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

		resp, _ := testRoleCreate(req, backend, t, "test_role1", data)
		if !resp.IsError() {
			t.Fatal("expecting error")
		}
		if errmsg, exp := resp.Data["error"].(string), "repo.repositories' field must be supplied"; !strings.Contains(errmsg, exp) {
			t.Errorf("err:%s exp:%#v\n", errmsg, exp)
		}

		if errmsg, exp := resp.Data["error"].(string), "repo.operations' field must be supplied"; !strings.Contains(errmsg, exp) {
			t.Errorf("err:%s exp:%#v\n", errmsg, exp)
		}
	})

	t.Run("permission_target_invalid_operation", func(t *testing.T) {

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
		resp, _ := testRoleCreate(req, backend, t, "test_role1", data)
		if !resp.IsError() {
			t.Fatal("expecting error")
		}
		if errmsg, exp := resp.Data["error"].(string), "operation 'invalidop' is not allowed"; !strings.Contains(errmsg, exp) {
			t.Errorf("err:%s exp:%#v\n", errmsg, exp)
		}
	})

}

func testRoleCreate(req *logical.Request, b logical.Backend, t *testing.T, roleName string, data map[string]interface{}) (*logical.Response, error) {

	req.Operation = logical.CreateOperation
	req.Path = fmt.Sprintf("roles/%s", roleName)
	req.Data = data

	resp, err := b.HandleRequest(context.Background(), req)
	return resp, err
}

func testRoleRead(req *logical.Request, b logical.Backend, t *testing.T, roleName string) (*logical.Response, error) {
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
	data := map[string]interface{}{
		"name": roleName,
	}

	req.Operation = logical.DeleteOperation
	req.Path = fmt.Sprintf("roles/%s", roleName)
	req.Data = data

	resp, err := b.HandleRequest(context.Background(), req)
	return resp, err
}

func testRoleList(req *logical.Request, b logical.Backend, t *testing.T) (*logical.Response, error) {

	req.Operation = logical.ListOperation
	req.Path = "roles"
	resp, err := b.HandleRequest(context.Background(), req)
	return resp, err
}

// Note: testing situation
// 1. role create
// 2. modification on exisitng permission target
// 3. check the permission target in artifactory
// 4. append a new permission target on top of existing ones
// 5. check newly permission target is created along with old ones in artifactory
// 6. remove a permission target
// 7. check if the permission target is removed from artifactory
// 8. remove a role
// 9. check artifactory group and permission targets are deleted in artifactory
