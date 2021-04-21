package artifactorysecrets

import (
	"context"
	"fmt"
	"net/http"

	"github.com/atlassian/go-artifactory/v2/artifactory"
	"github.com/atlassian/go-artifactory/v2/artifactory/transport"
	v1 "github.com/atlassian/go-artifactory/v2/artifactory/v1"
	v2 "github.com/atlassian/go-artifactory/v2/artifactory/v2"
	"github.com/hashicorp/vault/sdk/logical"
)

type ArtifactoryClient struct {
	client  *artifactory.Artifactory
	context context.Context
}

func (backend *ArtifactoryBackend) getArtifactoryClient(ctx context.Context, storage logical.Storage) (*ArtifactoryClient, error) {
	ac := &ArtifactoryClient{
		context: ctx,
	}
	config, err := backend.getConfig(ctx, storage)
	if err != nil {
		return ac, err
	}

	if config == nil {
		return ac, fmt.Errorf("artifactory backend configuration has not been set up")
	}

	c := &http.Client{} //nolint:ineffassign,staticcheck
	if config.BearerToken != "" {
		tp := transport.AccessTokenAuth{
			AccessToken: config.BearerToken,
		}
		c = tp.Client()
	} else if config.ApiKey != "" {
		tp := transport.ApiKeyAuth{
			ApiKey: config.ApiKey,
		}
		c = tp.Client()
	} else if config.Username != "" && config.Password != "" {
		tp := transport.BasicAuth{
			Username: config.Username,
			Password: config.Password,
		}
		c = tp.Client()
	} else {
		return ac, fmt.Errorf("bearer token, apikey or a pair of username/password isn't configured")
	}

	client, err := artifactory.NewClient(config.BaseURL, c)
	if err != nil {
		return ac, err
	}

	ac.client = client

	return ac, nil
}

func (ac *ArtifactoryClient) createOrReplaceGroup(role *RoleStorageEntry) (*http.Response, error) {
	name := groupName(role)
	desc := fmt.Sprintf("vault plugin group for %s", role.Name)
	group := v1.Group{
		Name:        &name,
		Description: &desc,
	}

	return ac.client.V1.Security.CreateOrReplaceGroup(ac.context, name, &group)
}

func (ac *ArtifactoryClient) deleteGroup(role *RoleStorageEntry) (*string, *http.Response, error) {
	return ac.client.V1.Security.DeleteGroup(ac.context, groupName(role))
}

func (ac *ArtifactoryClient) createOrUpdatePermissionTarget(role *RoleStorageEntry, pt *PermissionTarget) (*http.Response, error) {
	pt.Name = *permissionTargetName(role, pt.Name)
	cpt := &v2.PermissionTarget{}
	convertPermissionTarget(pt, cpt, role)

	exist, err := ac.client.V2.Security.HasPermissionTarget(ac.context, *cpt.Name)
	if err != nil {
		return nil, err
	}

	if exist {
		return ac.client.V2.Security.UpdatePermissionTarget(ac.context, *cpt.Name, cpt)
	}
	return ac.client.V2.Security.CreatePermissionTarget(ac.context, *cpt.Name, cpt)
}

func (ac *ArtifactoryClient) deletePermissionTarget(role *RoleStorageEntry, pt *PermissionTarget) (*http.Response, error) {
	ptName := *permissionTargetName(role, pt.Name)
	exist, err := ac.client.V2.Security.HasPermissionTarget(ac.context, ptName)
	if err != nil {
		return nil, err
	}
	if exist {
		return ac.client.V2.Security.DeletePermissionTarget(ac.context, ptName)
	}
	return nil, nil
}

func (ac *ArtifactoryClient) createToken(tokenReq TokenCreateEntry, role *RoleStorageEntry) (*v1.AccessToken, *http.Response, error) {

	u := tokenUsername(role.Name)
	ttlInSecond := int(tokenReq.TTL.Seconds())
	scope := fmt.Sprintf("member-of-groups:%s", groupName(role))
	acOpt := v1.AccessTokenOptions{
		Username:  &u,
		ExpiresIn: &ttlInSecond,
		Scope:     &scope,
	}

	return ac.client.V1.Security.CreateToken(ac.context, &acOpt)

}
