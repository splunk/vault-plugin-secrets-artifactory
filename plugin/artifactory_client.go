package artifactorysecrets

import (
	"context"
	"fmt"
	"net/http"

	"github.com/atlassian/go-artifactory/v2/artifactory"
	"github.com/atlassian/go-artifactory/v2/artifactory/transport"
	"github.com/hashicorp/vault/sdk/logical"
)

func (backend *ArtifactoryBackend) getArtifactoryClient(ctx context.Context, storage logical.Storage) (*artifactory.Artifactory, error) {
	client := &artifactory.Artifactory{}
	config, err := backend.getConfig(ctx, storage)
	if err != nil {
		return client, err
	}

	if config == nil {
		return client, fmt.Errorf("artifactory backend configuration has not been set up")
	}

	c := &http.Client{}
	if config.BearerToken != "" {
		tp := transport.AccessTokenAuth{
			AccessToken: config.BearerToken,
		}
		c = tp.Client()
	} else if config.Username != "" && config.Password != "" {
		tp := transport.BasicAuth{
			Username: config.Username,
			Password: config.Password,
		}
		c = tp.Client()
	} else {
		return client, fmt.Errorf("bearer token or a pair of username/password isn't configured")
	}

	client, err = artifactory.NewClient(config.BaseURL, c)
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return client, err
	}

	return client, nil
}
