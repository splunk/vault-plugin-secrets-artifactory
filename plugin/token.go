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
	"fmt"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
)

const (
	tokenPrefix = "token"
)

// TokenCreateEntry is the structure for creating a token
type TokenCreateEntry struct {
	TTL time.Duration `json:"ttl" structs:"ttl" mapstructure:"ttl"`
}

func (backend *ArtifactoryBackend) createTokenEntry(ctx context.Context, storage logical.Storage, createEntry TokenCreateEntry, roleEntry *RoleStorageEntry) (map[string]interface{}, error) {
	ac, err := backend.getClient(ctx, storage)
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
