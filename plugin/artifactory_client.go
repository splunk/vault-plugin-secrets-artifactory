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
	"io"
	"time"

	"github.com/jfrog/jfrog-client-go/access"
	accessauth "github.com/jfrog/jfrog-client-go/access/auth"
	accessservices "github.com/jfrog/jfrog-client-go/access/services"
	"github.com/jfrog/jfrog-client-go/artifactory"
	artauth "github.com/jfrog/jfrog-client-go/artifactory/auth"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/auth"
	artconfig "github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

const (
	clientTTL = 30 * time.Minute
)

type Client interface {
	CreateOrReplaceGroup(role *RoleStorageEntry) error
	DeleteGroup(role *RoleStorageEntry) error
	CreateOrUpdatePermissionTarget(role *RoleStorageEntry, pt *PermissionTarget, ptName string) error
	DeletePermissionTarget(ptName string) error
	CreateToken(tokenReq TokenCreateEntry, role *RoleStorageEntry) (auth.CreateTokenResponseData, error)
	Valid() bool
}

type artifactoryClient struct {
	client       artifactory.ArtifactoryServicesManager
	accessClient *access.AccessServicesManager

	expiration time.Time
}

var _ Client = &artifactoryClient{}

func NewClient(config *ConfigStorageEntry) (Client, error) {
	if config == nil {
		return nil, fmt.Errorf("artifactory backend configuration has not been set up")
	}

	ac := &artifactoryClient{
		expiration: time.Now().Add(clientTTL),
	}

	log.SetLogger(log.NewLogger(log.INFO, io.Discard))

	artifactoryDetails := artauth.NewArtifactoryDetails()
	artifactoryDetails.SetUrl(ensureArtifactoryURL(config.BaseURL))

	// For Access microservice
	accessDetails := accessauth.NewAccessDetails()
	accessDetails.SetUrl(ensureAccessURL(config.BaseURL))

	if config.BearerToken != "" {
		artifactoryDetails.SetAccessToken(config.BearerToken)
		accessDetails.SetAccessToken(config.BearerToken)
	} else if config.Username != "" && config.Password != "" {
		artifactoryDetails.SetUser(config.Username)
		artifactoryDetails.SetPassword(config.Password)
		accessDetails.SetUser(config.Username)
		accessDetails.SetPassword(config.Password)
	} else {
		return nil, fmt.Errorf("bearer token, or username/password isn't configured")
	}

	// Note: do not reuse Vault request context here as this client is cached between requests.
	artifactoryServiceConfig, err := artconfig.NewConfigBuilder().
		SetServiceDetails(artifactoryDetails).
		SetOverallRequestTimeout(config.ClientTimeout).
		// SetDryRun(false).
		SetContext(context.Background()).
		SetThreads(1).
		Build()

	if err != nil {
		return nil, fmt.Errorf("failed to build artifactory service config - %w", err)
	}

	client, err := artifactory.New(artifactoryServiceConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to build artifactory client- %w", err)
	}

	ac.client = client

	accessServicesConfig, err := artconfig.NewConfigBuilder().
		SetServiceDetails(accessDetails).
		SetContext(context.Background()).
		SetThreads(1).
		Build()

	if err != nil {
		return nil, fmt.Errorf("Failed to build access service config - %w", err)
	}

	accessClient, err := access.New(accessServicesConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to build access client - %w", err)
	}

	ac.accessClient = accessClient
	return ac, nil
}

func (ac *artifactoryClient) Valid() bool {
	return ac != nil && time.Now().Before(ac.expiration)
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
	*params.GroupDetails.AutoJoin = false
	*params.GroupDetails.AdminPrivileges = false
	return ac.client.CreateGroup(params)
}

func (ac *artifactoryClient) DeleteGroup(role *RoleStorageEntry) error {
	params := services.GroupParams{
		GroupDetails: services.Group{
			Name: groupName(role),
		},
	}
	group, err := ac.client.GetGroup(params)
	if err != nil {
		return err
	}
	if group != nil {
		return ac.client.DeleteGroup(group.Name)
	}
	return nil
}

func (ac *artifactoryClient) CreateOrUpdatePermissionTarget(role *RoleStorageEntry, pt *PermissionTarget, ptName string) error {
	params := services.PermissionTargetParams{}
	convertPermissionTarget(pt, &params, groupName(role), ptName)

	return ac.client.UpdatePermissionTarget(params)
}

func (ac *artifactoryClient) DeletePermissionTarget(ptName string) error {
	params, err := ac.client.GetPermissionTarget(ptName)
	if err != nil {
		return err
	}
	if params != nil {
		return ac.client.DeletePermissionTarget(params.Name)
	}
	return nil
}

func (ac *artifactoryClient) CreateToken(tokenReq TokenCreateEntry, role *RoleStorageEntry) (auth.CreateTokenResponseData, error) {
	expiresIn := uint(tokenReq.TTL.Seconds())
	params := accessservices.CreateTokenParams{
		CommonTokenParams: auth.CommonTokenParams{
			Scope:     fmt.Sprintf("applied-permissions/groups:%s", groupName(role)),
			ExpiresIn: &expiresIn,
			TokenType: "access_token",
			Audience:  "*@*",
		},
		Username:    tokenUsername(role.Name),
		Description: fmt.Sprintf("Generated from %s", pluginPrefix),
	}

	return ac.accessClient.CreateAccessToken(params)
}
