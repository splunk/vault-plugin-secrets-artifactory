package artifactorysecrets

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
)

// TokenCreateEntry is the structure for creating a token
type TokenCreateEntry struct {
	TTL  time.Duration `json:"ttl" structs:"ttl" mapstructure:"ttl"`
	Path string        `json:"path" structs:"path" mapstructure:"path"`
}

func (backend *ArtifactoryBackend) createTokenEntry(ctx context.Context, storage logical.Storage, createEntry TokenCreateEntry, roleEntry *RoleStorageEntry) (map[string]interface{}, error) {
	ac, err := backend.getArtifactoryClient(ctx, storage)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain artifactory client: %v", err)
	}

	token, _, err := ac.createToken(createEntry, roleEntry)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain a token: %v", err)
	}

	tokenOutput := map[string]interface{}{
		"access_token": token.AccessToken,
		"username":     tokenUsername(roleEntry.Name),
	}

	return tokenOutput, nil
}
