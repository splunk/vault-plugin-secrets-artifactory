package artifactorysecrets

import (
	"context"
	"fmt"
	"time"

	v1 "github.com/atlassian/go-artifactory/v2/artifactory/v1"
	"github.com/hashicorp/vault/sdk/logical"
)

// TokenCreateEntry is the structure for creating a token
type TokenCreateEntry struct {
	TTL  time.Duration `json:"ttl" structs:"ttl" mapstructure:"ttl"`
	Path string        `json:"path" structs:"path" mapstructure:"path"`
}

func (backend *ArtifactoryBackend) createTokenEntry(ctx context.Context, storage logical.Storage, createEntry TokenCreateEntry, roleEntry *RoleStorageEntry) (map[string]interface{}, error) {
	// TODO: create artifactory token with supplied group scope
	// --- super sloppy code below ---
	c, err := backend.getArtifactoryClient(ctx, storage)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain artifactory client: %v", err)
	}

	username := func(role_name string) string {
		return fmt.Sprintf("auto-vault-plugin.%s", role_name)
	}
	u := username(roleEntry.Name) //TODO: change
	ttlinsec := int(createEntry.TTL.Seconds())
	scope := fmt.Sprintf("member-of-groups:%s", groupName(roleEntry.RoleID))
	acOpt := v1.AccessTokenOptions{
		Username:  &u,
		ExpiresIn: &ttlinsec,
		Scope:     &scope,
	}

	token, _, err := c.V1.Security.CreateToken(ctx, &acOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to create a token: %v", err)
	}

	tokenOutput := map[string]interface{}{
		"access_token": token.AccessToken,
		"username":     u,
	}

	// --- super sloppy code above ---

	return tokenOutput, nil
}
