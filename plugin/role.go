package artifactorysecrets

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
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
func (backend *ArtifactoryBackend) saveRoleWithNewPermissionTargets(ctx context.Context, req *logical.Request, role *RoleStorageEntry, pts []PermissionTarget) (warning []string, err error) {
	backend.Logger().Debug("Creating/Updating role with new permission targets")

	oldPts := role.PermissionTargets

	cfg, err := backend.getConfig(ctx, req.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain artifactory config - %s", err.Error())
	}

	ac, err := backend.getClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain artifactory client - %s", err.Error())
	}

	// Create/update a group
	backend.Logger().Debug("creating/updating a group", "name", role.Name, "role_id", role.RoleID)
	if err := ac.CreateOrReplaceGroup(role); err != nil {
		return nil, fmt.Errorf("failed to create an artifactory group - %s", err.Error())
	}

	if len(oldPts) > len(pts) {
		backend.Logger().Debug("removing role excessive permission targets", "role_name", role.Name)
		if cleanupErr := backend.tryDeleteRoleResources(ctx, req, role, oldPts[len(pts):], len(pts), false); cleanupErr != nil {
			backend.Logger().Warn(
				"unable to clean up unused old permission targets for role.",
				"role_name", role.Name, "errors", cleanupErr)
			return []string{cleanupErr.Error()}, nil
		}
	}

	// update permission target in role before save
	role.PermissionTargets = pts
	if err = role.save(ctx, req.Storage); err != nil {
		return nil, err
	}

	// Create/Update permission targets
	for idx, pt := range pts {
		ptName := permissionTargetName(role.Name, idx)
		backend.Logger().Debug("creating/updating a permission target", "name", ptName)
		if err := ac.CreateOrUpdatePermissionTarget(role, &pt, ptName); err != nil {
			return nil, fmt.Errorf("Failed to create/update a permission target - %s", err.Error())
		}
	}

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

func (backend *ArtifactoryBackend) tryDeleteRoleResources(ctx context.Context, req *logical.Request, role *RoleStorageEntry, pts []PermissionTarget, offset int, deleteGroup bool) error {
	if len(pts) == 0 {
		backend.Logger().Debug("skip deletion for empty permission targets")
	}

	cfg, err := backend.getConfig(ctx, req.Storage)
	if err != nil {
		return fmt.Errorf("failed to obtain artifactory config - %s", err.Error())
	}

	ac, err := backend.getClient(ctx, cfg)
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
