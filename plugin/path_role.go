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
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// schema for the creation of the role, this will map the fields coming in from the
// vault request field map
var createRoleSchema = map[string]*framework.FieldSchema{
	"name": {
		Type:        framework.TypeString,
		Description: "The name of the role to be created",
	},
	"token_ttl": {
		Type:        framework.TypeDurationSecond,
		Description: "The TTL of the token",
		Default:     900,
	},
	"max_ttl": {
		Type:        framework.TypeDurationSecond,
		Description: "The TTL of the token",
		Default:     3600,
	},
	"permission_targets": {
		Type:        framework.TypeString,
		Description: "List of permission target configurations",
	},
}

// remove the specified role from the storage
func (backend *ArtifactoryBackend) pathRoleDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	roleName := data.Get("name").(string)
	if roleName == "" {
		return logical.ErrorResponse("Unable to remove, missing role name"), nil
	}

	lock := backend.roleLock(roleName)
	lock.RLock()
	defer lock.RUnlock()

	// get the role to make sure it exists and to get the role id
	role, err := getRoleEntry(ctx, req.Storage, roleName)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, nil
	}

	deleteGroup := true

	if err := backend.deleteRoleEntry(ctx, req.Storage, roleName); err != nil {
		return logical.ErrorResponse(fmt.Sprintf("Unable to remove role %s", roleName)), err
	}

	// Try to clean up resources.
	if cleanupErr := backend.tryDeleteRoleResources(ctx, req, role, role.PermissionTargets, 0, deleteGroup); cleanupErr != nil {
		backend.Logger().Warn(
			"unable to clean up unused artifactory resources from deleted role.",
			"role_name", roleName, "errors", cleanupErr)
		return &logical.Response{Warnings: []string{cleanupErr.Error()}}, nil
	}

	backend.Logger().Debug("successfully deleted role and artifactory resources", "name", roleName)
	return nil, nil
}

// read the current role from the inputs and return it if it exists
func (backend *ArtifactoryBackend) pathRoleRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	roleName := data.Get("name").(string)
	role, err := getRoleEntry(ctx, req.Storage, roleName)
	if err != nil {
		return logical.ErrorResponse("Error reading role"), err
	}

	if role == nil {
		return nil, nil
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"name":               role.Name,
			"id":                 role.RoleID,
			"token_ttl":          int64(role.TokenTTL / time.Second),
			"max_ttl":            int64(role.MaxTTL / time.Second),
			"permission_targets": role.RawPermissionTargets,
		},
	}, nil
}

// read the current role from the inputs and return it if it exists
func (backend *ArtifactoryBackend) pathRolesList(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	roles, err := backend.listRoleEntries(ctx, req.Storage)
	if err != nil {
		return logical.ErrorResponse("Error listing roles"), err
	}
	return logical.ListResponse(roles), nil
}

func (backend *ArtifactoryBackend) pathRoleCreateUpdate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	roleDetails := func(role *RoleStorageEntry) map[string]interface{} {
		return map[string]interface{}{
			"role_id":            role.RoleID,
			"role_name":          role.Name,
			"permission_targets": role.RawPermissionTargets,
		}
	}

	roleName := data.Get("name").(string)
	if roleName == "" {
		return logical.ErrorResponse("Role name not supplied"), nil
	}

	lock := backend.roleLock(roleName)
	lock.RLock()
	defer lock.RUnlock()

	config, err := backend.getConfig(ctx, req.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain artifactory config - %s", err.Error())
	}
	if config == nil {
		return nil, fmt.Errorf("artifactory backend configuration has not been set up")
	}

	role, err := getRoleEntry(ctx, req.Storage, roleName)
	if err != nil {
		return logical.ErrorResponse("Error reading role"), nil
	}

	if role == nil {
		role = &RoleStorageEntry{
			Name: roleName,
		}
		roleID, _ := uuid.NewUUID()
		role.RoleID = roleID.String()
	}

	isCreate := req.Operation == logical.CreateOperation

	// Permission Targets
	ptsRaw, newPermissionTargets := data.GetOk("permission_targets")
	if newPermissionTargets {
		pts, ok := ptsRaw.(string)
		if !ok {
			return logical.ErrorResponse("permission targets are not a string"), nil
		}
		if pts == "" {
			return logical.ErrorResponse("permission targets are empty"), nil
		}
	}

	if isCreate && !newPermissionTargets {
		return logical.ErrorResponse("permission targets are required for new role"), nil
	}

	maxttlRaw, ok := data.GetOk("max_ttl")
	if ok && maxttlRaw.(int) > 0 {
		role.MaxTTL = time.Duration(maxttlRaw.(int)) * time.Second
	} else {
		role.MaxTTL = time.Duration(createRoleSchema["max_ttl"].Default.(int)) * time.Second
	}

	ttlRaw, ok := data.GetOk("token_ttl")
	if ok && ttlRaw.(int) > 0 {
		role.TokenTTL = time.Duration(ttlRaw.(int)) * time.Second
	} else {
		role.TokenTTL = time.Duration(createRoleSchema["token_ttl"].Default.(int)) * time.Second
	}

	if role.MaxTTL > config.MaxTTL {
		return logical.ErrorResponse(fmt.Sprintf("role max ttl is greater than config max ttl '%d'", config.MaxTTL)), nil
	}
	if role.TokenTTL > role.MaxTTL {
		return logical.ErrorResponse(fmt.Sprintf("role token ttl is greater than role max ttl '%d'", role.MaxTTL)), nil
	}
	// If no new permission targets or new permission targets are exactly same as old permission targets,
	// just return without updating permission targets
	// if !newPermissionTargets || role.permissionTargetsHash() == getStringHash(ptsRaw.(string)) {
	// 	backend.Logger().Debug("No net new permission targets are added for role", "role_name", role.Name)
	// 	if err := role.save(ctx, req.Storage); err != nil {
	// 		return logical.ErrorResponse(err.Error()), nil
	// 	}
	// 	return &logical.Response{Data: roleDetails(role)}, nil
	// }

	// new permission targets, update role
	var pts []PermissionTarget
	err = json.Unmarshal([]byte(ptsRaw.(string)), &pts)
	if err != nil {
		return logical.ErrorResponse("Error unmarshal permission targets. Expecting list of permission targets - " + err.Error()), nil
	}
	if len(pts) == 0 {
		return logical.ErrorResponse("Failed to parse any permission targets from given permission targets JSON"), nil
	}
	for _, pt := range pts {
		if err = pt.assertValid(); err != nil {
			return logical.ErrorResponse("Failed to validate a permission target - " + err.Error()), nil
		}
	}
	role.RawPermissionTargets = ptsRaw.(string)

	// save role with new permission targets
	warnings, err := backend.saveRoleWithNewPermissionTargets(ctx, req, role, pts)
	if err != nil {
		return logical.ErrorResponse(err.Error()), nil
	} else if len(warnings) > 0 {
		return &logical.Response{Warnings: warnings, Data: roleDetails(role)}, nil
	}

	return &logical.Response{Data: roleDetails(role)}, nil
}

func (backend *ArtifactoryBackend) pathRoleExistenceCheck(roleFieldName string) framework.ExistenceFunc {
	return func(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
		roleName := data.Get(roleFieldName).(string)
		role, err := getRoleEntry(ctx, req.Storage, roleName)
		if err != nil {
			return false, err
		}

		return role != nil, nil
	}
}

// set up the paths for the roles within vault
func pathRole(backend *ArtifactoryBackend) []*framework.Path {
	paths := []*framework.Path{
		{
			Pattern:        fmt.Sprintf("%s/%s", rolesPrefix, framework.GenericNameRegex("name")),
			Fields:         createRoleSchema,
			ExistenceCheck: backend.pathRoleExistenceCheck("name"),
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.CreateOperation: backend.pathRoleCreateUpdate,
				logical.UpdateOperation: backend.pathRoleCreateUpdate,
				logical.ReadOperation:   backend.pathRoleRead,
				logical.DeleteOperation: backend.pathRoleDelete,
			},
			HelpSynopsis:    pathRoleHelpSyn,
			HelpDescription: pathRoleHelpDesc,
		},
	}

	return paths
}

func pathRoleList(backend *ArtifactoryBackend) []*framework.Path {
	// Paths for listing role sets
	paths := []*framework.Path{
		{
			Pattern: fmt.Sprintf("%s?/?", rolesPrefix),
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.ListOperation: backend.pathRolesList,
			},
			HelpSynopsis: pathListRoleHelpSyn,
		},
	}
	return paths
}

const pathRoleHelpSyn = `Read/write sets of permission targets to be given to generated credentials for specified role.`
const pathRoleHelpDesc = `
This path allows you to create roles, which bind sets of permission targets
of specific repositories with patterns and operations to a group. Secrets are 
generated under a role and will have the given set of permission targets on group.

The specified permission targets file accepts an JSON string
with the following format:

[
  {
    "repo": {
      "include_patterns": ["**"] (default),
      "exclude_patterns": [""] (default),
      "repositories": ["local-repo1", "local-repo2", "remote-repo1", "virtual-repo2"],
      "operations": ["read"]
    },
    "build": {
      "include_patterns": ["**"] (default),
      "exclude_patterns": [""] (default),
      "repositories": ["artifactory-build-info"], (default, can't be changed)
      "operations": ["manage","read","annotate"]
    },
  }
]

At least one of repo or build is required

| field | subfield         | required |
| ----- | ---------------- | -------- |
| repo  | N/A              | no       | 
|       | include_patterns | no       | 
|       | exclude_patterns | no       | 
|       | repositories     | yes      | 
|       | operations       | yes      | 
| build | N/A              | no       | 
|       | include_patterns | no       | 
|       | exclude_patterns | no       | 
|       | repositories     | yes      | 
|       | operations       | yes      |

Allowed operations are "read", "write", "annotate",
"delete", "manage", "managedXrayMeta", "distribute"
`

const pathListRoleHelpSyn = `List existing roles.`
