package artifactorysecrets

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/mitchellh/mapstructure"
)

func TestPathRole(t *testing.T) {
	b, storage := getTestBackend(t)

	conf := map[string]interface{}{
		"base_url":     "https://example.jfrog.io/example",
		"bearer_token": "mybearertoken",
		"max_ttl":      "600s",
	}

	testConfigUpdate(t, b, storage, conf)
	// TODO: need to mock artifactory client so that test doesn't necessarily hit artifactory intance

	/***  TEST CREATE OPERATION ***/
	req := &logical.Request{
		Storage: storage,
	}

	resp, err := testRoleCreate(req, b, t, "test_role1")

	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%s resp:%#v\n", err, resp)
	}

	resp, err = testRoleCreate(req, b, t, "test_role2")

	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%s resp:%#v\n", err, resp)
	}

	/***  TEST GET OPERATION ***/
	resp, err = testRoleRead(req, b, t, "test_role1")
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%s resp:%#v\n", err, resp)
	}

	var returnedRole RoleStorageEntry
	err = mapstructure.Decode(resp.Data, &returnedRole)

	if returnedRole.Name != "test_role1" {
		t.Fatalf("incorrect role name %s returned, not the same as saved value \n", returnedRole.Name)
	}

	/*** TEST GET NON-EXISTENT ROLE ***/
	resp, err = testRoleRead(req, b, t, "test_role3")
	if err != nil && resp != nil {
		t.Fatalf("err:%s resp:%#v\n", err, resp)
	}

	/***  TEST List OPERATION ***/
	resp, err = testRoleList(req, b, t)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%s resp:%#v\n", err, resp)
	}

	var listResp map[string]interface{}
	err = mapstructure.Decode(resp.Data, &listResp)

	returnedRoles := listResp["keys"].([]string)

	if len(returnedRoles) != 2 {
		t.Fatalf("incorrect number of roles \n")
	}

	if returnedRoles[0] != "test_role1" && returnedRoles[1] != "test_role2" {
		t.Fatalf("incorrect path set \n")
	}

	/***  TEST Delete OPERATION ***/
	resp, err = testRoleDelete(req, b, t, "test_role1")
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%s resp:%#v\n", err, resp)
	}
}

func testRoleCreate(req *logical.Request, b logical.Backend, t *testing.T, roleName string) (*logical.Response, error) {
	data := map[string]interface{}{
		"name": roleName,
	}

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
	req.Path = fmt.Sprintf("roles")
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
// 9. check artifactory group and permissino targets are deleted in artifactory
