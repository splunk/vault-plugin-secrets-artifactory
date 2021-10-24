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
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	rolesPrefix = "roles"
)

type RoleStorageEntry struct {
	// `json:"" structs:"" mapstructure:""`
	// The UUID that defines this role
	RoleID string `json:"role_id" structs:"role_id" mapstructure:"role_id"`

	// The TTL for your token
	TokenTTL time.Duration `json:"token_ttl" structs:"token_ttl" mapstructure:"token_ttl"`

	// The Max TTL for your token
	MaxTTL time.Duration `json:"max_ttl" structs:"max_ttl" mapstructure:"max_ttl"`

	// The provided name for the role
	Name string `json:"name" structs:"name" mapstructure:"name"`

	RawPermissionTargets string
	PermissionTargets    []PermissionTarget
}

// validate checks whether a Role has been populated properly before saving
func (role RoleStorageEntry) validate() error {
	var err *multierror.Error
	if role.Name == "" {
		err = multierror.Append(err, errors.New("role name is empty"))
	}
	if role.RoleID == "" {
		err = multierror.Append(err, errors.New("role id is empty"))
	}
	if role.RawPermissionTargets == "" {
		err = multierror.Append(err, errors.New("raw permission targets are empty"))
	}
	if role.PermissionTargets == nil {
		err = multierror.Append(err, errors.New("permission targets are empty"))
	}
	return err.ErrorOrNil()
}

// save saves a role to storage
func (role RoleStorageEntry) save(ctx context.Context, storage logical.Storage) error {
	if err := role.validate(); err != nil {
		return err
	}

	entry, err := logical.StorageEntryJSON(fmt.Sprintf("%s/%s", rolesPrefix, role.Name), role)
	if err != nil {
		return err
	}

	return storage.Put(ctx, entry)
}

func (role RoleStorageEntry) permissionTargetsHash() string {
	return getStringHash(role.RawPermissionTargets)
}

// get or create the basic lock for the role name
func (backend *ArtifactoryBackend) roleLock(roleName string) *locksutil.LockEntry {
	return locksutil.LockForKey(backend.roleLocks, roleName)
}

// saveRoleWithNewPermissionTargets will create group and permission targets
// persist in the data store
func (backend *ArtifactoryBackend) saveRoleWithNewPermissionTargets(ctx context.Context, req *logical.Request, role *RoleStorageEntry, pts []PermissionTarget, deleteGroup bool) (warning []string, err error) {
	backend.Logger().Debug("Creating/Updating role with new permission targets")

	oldPts := role.PermissionTargets

	// Add WALs for both old and new permission targets
	// WAL callback checks whether resources are still being used by role
	// so there is no harm in adding WALs early, or adding WALs for resources that
	// will eventually get cleaned up
	deleteExcessivePTs := len(oldPts) > len(pts)
	oldWalIds := []string{}
	if deleteExcessivePTs {
		backend.Logger().Debug("adding WALs for old permission targets")
		oldWalIds, err = backend.addWalsForRoleResources(ctx, req, role.Name, oldPts[len(pts):], len(pts), false)
		if err != nil {
			return nil, err
		}
	}

	backend.Logger().Debug("adding WALs for new permission targets")
	newWalIds, err := backend.addWalsForRoleResources(ctx, req, role.Name, pts, 0, deleteGroup)
	if err != nil {
		return nil, err
	}

	ac, err := backend.getClient(ctx, req.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain artifactory client - %s", err.Error())
	}

	// Create/update a group
	backend.Logger().Debug("creating/updating a group", "name", role.Name, "role_id", role.RoleID)
	if err := ac.CreateOrReplaceGroup(role); err != nil {
		return nil, fmt.Errorf("failed to create an artifactory group - %s", err.Error())
	}

	if deleteExcessivePTs {
		backend.Logger().Debug("removing role excessive permission targets", "role_name", role.Name)
		if cleanupErr := backend.tryDeleteRoleResources(ctx, req, role, oldPts[len(pts):], len(pts), false, oldWalIds); cleanupErr != nil {
			backend.Logger().Warn(
				"unable to clean up unused old permission targets for role. WALs exist to cleanup but ignoring error",
				"role_name", role.Name, "errors", cleanupErr)
			return []string{cleanupErr.Error()}, nil
		}
	}

	// Create/Update permission targets
	for idx, pt := range pts {
		ptName := permissionTargetName(role.Name, idx)
		backend.Logger().Debug("creating/updating a permission target", "name", ptName)
		if err := ac.CreateOrUpdatePermissionTarget(role, &pt, ptName); err != nil {
			return nil, fmt.Errorf("Failed to create/update a permission target - %s", err.Error())
		}
	}

	// update permission target in role before save
	role.PermissionTargets = pts
	if err = role.save(ctx, req.Storage); err != nil {
		return nil, err
	}

	// Successfully saved the new role with new permission targets. try cleaning up WALs
	// that would rollback the role permission targets (will no-op if still in use by role)
	backend.Logger().Debug("removing WALs for new permission targets")
	backend.tryDeleteWALs(ctx, req.Storage, newWalIds...)

	return nil, nil
}

// deleteRoleEntry will remove the role with specified name from storage
func (backend *ArtifactoryBackend) deleteRoleEntry(ctx context.Context, storage logical.Storage, roleName string) error {
	if roleName == "" {
		return fmt.Errorf("missing role name")
	}

	return storage.Delete(ctx, fmt.Sprintf("%s/%s", rolesPrefix, roleName))
}

// getRoleEntry fetches a role from the storage
func getRoleEntry(ctx context.Context, storage logical.Storage, roleName string) (*RoleStorageEntry, error) {
	var result RoleStorageEntry
	if entry, err := storage.Get(ctx, fmt.Sprintf("%s/%s", rolesPrefix, roleName)); err != nil {
		return nil, err
	} else if entry == nil {
		return nil, nil
	} else if err := entry.DecodeJSON(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// listRoleEntries gets all the roles
func (backend *ArtifactoryBackend) listRoleEntries(ctx context.Context, storage logical.Storage) ([]string, error) {
	roles, err := storage.List(ctx, fmt.Sprintf("%s/", rolesPrefix))
	if err != nil {
		return nil, err
	}
	return roles, nil
}
func (backend *ArtifactoryBackend) addWalsForRoleResources(ctx context.Context, req *logical.Request, roleName string, pts []PermissionTarget, offset int, deleteGroup bool) ([]string, error) {
	if len(pts) == 0 {
		backend.Logger().Debug("skip WALs for nil role permission targets")
		return nil, nil
	}
	walIds := make([]string, 0, len(pts)+2)
	var err error
	var walId string
	if deleteGroup {
		backend.Logger().Debug("adding WAL for group", "role_name", roleName)
		walId, err = framework.PutWAL(ctx, req.Storage, walTypeGroup, &walGroup{
			RoleName: roleName,
		})
		if err != nil {
			return walIds, errwrap.Wrapf("Failed to create WAL entry to clean up group: {{err}}", err)
		}
		walIds = append(walIds, walId)
	}

	for idx := range pts {
		ptName := permissionTargetName(roleName, idx+offset)
		backend.Logger().Debug("adding WAL for permission target", "name", ptName)
		walId, err = framework.PutWAL(ctx, req.Storage, walTypePermissionTarget, &walPermissionTarget{
			RoleName:             roleName,
			PermissionTargetName: ptName,
			Index:                idx + offset,
		})
		if err != nil {
			return walIds, errwrap.Wrapf("Failed to create WAL entry to clean up a permission target: {{err}}", err)
		}
		walIds = append(walIds, walId)
	}

	return walIds, err
}

func (backend *ArtifactoryBackend) tryDeleteRoleResources(ctx context.Context, req *logical.Request, role *RoleStorageEntry, pts []PermissionTarget, offset int, deleteGroup bool, walIds []string) error {
	if len(pts) == 0 {
		backend.Logger().Debug("skip deletion for empty permission targets")
	}

	ac, err := backend.getClient(ctx, req.Storage)
	if err != nil {
		return fmt.Errorf("failed to obtain artifactory client - %s", err.Error())
	}

	var merr *multierror.Error

	if deleteGroup {
		if err = ac.DeleteGroup(role); err != nil {
			backend.Logger().Info("Deleting group from artifactory", "name", groupName(role), "role", role.Name)
			merr = multierror.Append(merr, fmt.Errorf("failed to delete a group for role %s - %s", role.Name, err.Error()))
		}
	}

	for idx := range pts {
		ptName := permissionTargetName(role.Name, idx+offset)
		backend.Logger().Info("Deleting permission target from artifactory", "name", ptName, "role_name", role.Name)
		if err := ac.DeletePermissionTarget(ptName); err != nil {
			merr = multierror.Append(merr, fmt.Errorf("failed to delete a permission target %s for role %s - %s", ptName, role.Name, err.Error()))
		}
	}

	return merr.ErrorOrNil()
}

func (backend *ArtifactoryBackend) tryDeleteWALs(ctx context.Context, storage logical.Storage, walIds ...string) {
	for _, walId := range walIds {
		// ignore errors, WALs that are not needed will just no-op
		err := framework.DeleteWAL(ctx, storage, walId)
		if err != nil {
			backend.Logger().Error("Unable to delete unneeded WAL %s, ignoring error since WAL will no-op: %v", walId, err)
		}
	}
}
