package artifactorysecrets

import (
	"context"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
)

const (
	configPrefix = "config"
)

// ConfigStorageEntry structure represents the config as it is stored within vault
type ConfigStorageEntry struct {
	BaseURL     string        `json:"base_url" structs:"base_url" mapstructure:"base_url"`
	BearerToken string        `json:"bearer_token" structs:"bearer_token" mapstructure:"bearer_token"`
	Username    string        `json:"username" structs:"username" mapstructure:"username"`
	Password    string        `json:"password" structs:"password" mapstructure:"password"`
	ApiKey      string        `json:"api_key" structs:"api_key" mapstructure:"api_key"`
	MaxTTL      time.Duration `json:"max_ttl" structs:"max_ttl" mapstructure:"max_ttl"`
}

func (backend *ArtifactoryBackend) getConfig(ctx context.Context, s logical.Storage) (*ConfigStorageEntry, error) {
	var cfg ConfigStorageEntry
	cfgRaw, err := s.Get(ctx, configPrefix)
	if err != nil {
		return nil, err
	}
	if cfgRaw == nil {
		return nil, nil
	}

	if err := cfgRaw.DecodeJSON(&cfg); err != nil {
		return nil, err
	}

	return &cfg, err
}
