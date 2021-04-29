package artifactorysecrets

import (
	"context"
	"fmt"
	"strings"

	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/auth"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	artconfig "github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type Client interface {
	CreateOrReplaceGroup(role *RoleStorageEntry) error
	DeleteGroup(role *RoleStorageEntry) error
	CreateOrUpdatePermissionTarget(role *RoleStorageEntry, pt *PermissionTarget, ptName string) error
	DeletePermissionTarget(ptName string) error
	CreateToken(tokenReq TokenCreateEntry, role *RoleStorageEntry) (services.CreateTokenResponseData, error)
}

type artifactoryClient struct {
	client  artifactory.ArtifactoryServicesManager
	context context.Context
}

var _ Client = &artifactoryClient{}

func NewClient(ctx context.Context, config *ConfigStorageEntry) (Client, error) {
	if config == nil {
		return nil, fmt.Errorf("artifactory backend configuration has not been set up")
	}

	ac := &artifactoryClient{
		context: ctx,
	}

	// TODO: need to figure out
	log.SetLogger(log.NewLogger(log.INFO, nil))

	artifactoryDetails := auth.NewArtifactoryDetails()
	artifactoryDetails.SetUrl(config.BaseURL)

	if config.BearerToken != "" {
		artifactoryDetails.SetAccessToken(config.BearerToken)
	} else if config.ApiKey != "" {
		artifactoryDetails.SetApiKey(config.ApiKey)
	} else if config.Username != "" && config.Password != "" {
		artifactoryDetails.SetUser(config.Username)
		artifactoryDetails.SetPassword(config.Password)
	} else {
		return nil, fmt.Errorf("bearer token, apikey or a pair of username/password isn't configured")
	}

	artifactoryServiceConfig, err := artconfig.NewConfigBuilder().
		SetServiceDetails(artifactoryDetails).
		// SetDryRun(false).
		SetContext(ctx).
		SetThreads(1).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build artifactory service config - %v", err.Error())
	}

	client, err := artifactory.New(artifactoryServiceConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to build artifactory client- %v", err.Error())
	}
	ac.client = client

	return ac, nil
}

func (ac *artifactoryClient) CreateOrReplaceGroup(role *RoleStorageEntry) error {
	params := services.GroupParams{
		GroupDetails: services.Group{
			Name: groupName(role),
		},
	}

	group, err := ac.client.GetGroup(params)
	if err != nil {
		return fmt.Errorf("Error fetching a group '%s' - %s", groupName(role), err)
	}
	if group != nil {
		params.ReplaceIfExists = true
		params.GroupDetails = *group
		return ac.client.UpdateGroup(params)
	}
	params.GroupDetails.Description = fmt.Sprintf("vault plugin group for %s", role.Name)
	params.GroupDetails.AutoJoin = false
	params.GroupDetails.AdminPrivileges = false
	return ac.client.CreateGroup(params)
}

func (ac *artifactoryClient) DeleteGroup(role *RoleStorageEntry) error {
	return ac.client.DeleteGroup(groupName(role))
}

func (ac *artifactoryClient) CreateOrUpdatePermissionTarget(role *RoleStorageEntry, pt *PermissionTarget, ptName string) error {
	// params, err := ac.client.GetPermissionTarget(ptName)
	// if ignoreNotFound(err) != nil {
	// 	return err
	// }

	// if params != nil {
	// 	// update logic
	// }
	// return ac.client.CreatePermissionTarget(*params)
	params := services.PermissionTargetParams{}
	convertPermissionTarget(pt, &params, groupName(role), ptName)

	return ac.client.UpdatePermissionTarget(params)
}

func (ac *artifactoryClient) DeletePermissionTarget(ptName string) error {
	params, err := ac.client.GetPermissionTarget(ptName)
	if ignoreNotFound(err) != nil {
		return err
	}
	if params != nil {
		return ac.client.DeletePermissionTarget(params.Name)
	}
	return nil
}

func (ac *artifactoryClient) CreateToken(tokenReq TokenCreateEntry, role *RoleStorageEntry) (services.CreateTokenResponseData, error) {
	params := services.CreateTokenParams{
		Scope:     fmt.Sprintf("api:* member-of-groups:%s", groupName(role)),
		Username:  tokenUsername(role.Name),
		ExpiresIn: int(tokenReq.TTL.Seconds()),
	}

	return ac.client.CreateToken(params)
}

// This is temporary until Get Permission Target API returns nil err in case NotFound
// https://github.com/jfrog/jfrog-client-go/pull/337
func ignoreNotFound(err error) error {
	if err == nil {
		return err
	}
	// API error in case status is not 200 OK
	// "Artifactory response: " + resp.Status + "yadayadaya"
	notFoundStatus := 404
	if strings.Contains(err.Error(), fmt.Sprintf("Artifactory response: %d", notFoundStatus)) {
		return nil
	}
	return err
}
