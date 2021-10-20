package artifactorysecrets

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/mitchellh/mapstructure"
)

const (
	walTypeGroup            = "group"
	walTypePermissionTarget = "permission_target"
)

func (backend *ArtifactoryBackend) walRollback(ctx context.Context, req *logical.Request, kind string, data interface{}) error {
	switch kind {
	case walTypeGroup:
		return backend.groupRollback(ctx, req, data)
	case walTypePermissionTarget:
		return backend.permissionTargetRollback(ctx, req, data)
	default:
		return fmt.Errorf("unknown type to rollback")
	}
}

type walGroup struct {
	RoleName string
	ID       string
}

type walPermissionTarget struct {
	RoleName             string
	PermissionTargetName string
	Index                int
}

func (backend *ArtifactoryBackend) groupRollback(ctx context.Context, req *logical.Request, data interface{}) error {

	var entry walGroup
	if err := mapstructure.Decode(data, &entry); err != nil {
		return err
	}

	if entry.RoleName == "" {
		return fmt.Errorf("missing role name")
	}

	lock := backend.roleLock(entry.RoleName)
	lock.RLock()
	defer lock.RUnlock()

	// If group is still being used, WAL entry was not deleted properly
	// after a successful operations. Remove WAL entry
	role, err := getRoleEntry(ctx, req.Storage, entry.RoleName)
	if err != nil {
		return err
	}
	if role != nil && entry.ID == role.RoleID {
		// still being used - don't delete this group
		return nil
	}

	// delete group
	ac, err := backend.getClient(ctx, req.Storage)
	if err != nil {
		return fmt.Errorf("failed to obtain artifactory client - %s", err.Error())
	}

	backend.Logger().Debug("Deleting group by WAL rollback", "name", entry.RoleName)
	return backend.deleteGroup(ctx, ac, entry.ID)
}

func (backend *ArtifactoryBackend) permissionTargetRollback(ctx context.Context, req *logical.Request, data interface{}) error {
	var entry walPermissionTarget
	if err := mapstructure.Decode(data, &entry); err != nil {
		return err
	}

	if entry.RoleName == "" {
		return fmt.Errorf("missing role name")
	}

	lock := backend.roleLock(entry.RoleName)
	lock.RLock()
	defer lock.RUnlock()

	// If group is still being used, WAL entry was not deleted properly
	// after a successful operations. Remove WAL entry
	role, err := getRoleEntry(ctx, req.Storage, entry.RoleName)
	if err != nil {
		return err
	}
	if role != nil && entry.Index > len(role.PermissionTargets) {
		// Still being used
		return nil
	}

	// delete permission target
	ac, err := backend.getClient(ctx, req.Storage)
	if err != nil {
		return fmt.Errorf("failed to obtain artifactory client - %s", err.Error())
	}

	backend.Logger().Debug("Deleting permission target by WAL rollback", "name", entry.PermissionTargetName)
	return backend.deletePermissionTarget(ctx, ac, entry.PermissionTargetName)
}

func (backend *ArtifactoryBackend) deleteGroup(ctx context.Context, ac Client, roleID string) error {
	if roleID == "" {
		return nil
	}

	roleEntry := RoleStorageEntry{RoleID: roleID}
	return ac.DeleteGroup(&roleEntry)
}

func (backend *ArtifactoryBackend) deletePermissionTarget(ctx context.Context, ac Client, ptName string) error {
	if ptName == "" {
		return nil
	}

	return ac.DeletePermissionTarget(ptName)
}
