package artifactorysecrets

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
)

// TokenCreateEntry is the structure for creating a token
type TokenCreateEntry struct {
	TTL time.Duration `json:"ttl" structs:"ttl" mapstructure:"ttl"`
}

func (backend *ArtifactoryBackend) createTokenEntry(ctx context.Context, storage logical.Storage, createEntry TokenCreateEntry, roleEntry *RoleStorageEntry) (map[string]interface{}, error) {
	cfg, err := backend.getConfig(ctx, storage)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain artifactory config: %v", err)
	}

	ac, err := backend.getClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain artifactory client: %v", err)
	}

	token, err := ac.CreateToken(createEntry, roleEntry)
	if err != nil {
		return nil, fmt.Errorf("failed to create a token: %v", err)
	}

	tokenOutput := map[string]interface{}{
		"access_token": token.AccessToken,
		"username":     tokenUsername(roleEntry.Name),
	}

	return tokenOutput, nil
}
