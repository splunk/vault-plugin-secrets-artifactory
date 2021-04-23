package artifactorysecrets

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

// get or create the basic lock for the role name
func (backend *ArtifactoryBackend) roleLock(roleName string) *locksutil.LockEntry {
	return locksutil.LockForKey(backend.roleLocks, roleName)
}

// roleSave will persist the role in the data store
func (backend *ArtifactoryBackend) setRoleEntry(ctx context.Context, storage logical.Storage, role RoleStorageEntry) error {
	if role.Name == "" {
		return fmt.Errorf("Unable to save, invalid name in role")
	}

	roleName := strings.ToLower(role.Name)

	lock := backend.roleLock(roleName)
	lock.RLock()
	defer lock.RUnlock()

	return role.save(ctx, storage)
}

// deleteRoleEntry will remove the role with specified name from storage
func (backend *ArtifactoryBackend) deleteRoleEntry(ctx context.Context, storage logical.Storage, roleName string) error {
	if roleName == "" {
		return fmt.Errorf("missing role name")
	}
	roleName = strings.ToLower(roleName)

	lock := backend.roleLock(roleName)
	lock.RLock()
	defer lock.RUnlock()

	return storage.Delete(ctx, fmt.Sprintf("%s/%s", rolesPrefix, roleName))
}

// getRoleEntry fetches a role from the storage
func (backend *ArtifactoryBackend) getRoleEntry(ctx context.Context, storage logical.Storage, roleName string) (*RoleStorageEntry, error) {
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
	roles, err := storage.List(ctx, "roles/")
	if err != nil {
		return nil, err
	}
	return roles, nil
}
