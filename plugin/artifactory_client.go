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

	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/auth"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
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
	CreateToken(tokenReq TokenCreateEntry, role *RoleStorageEntry) (services.CreateTokenResponseData, error)
	Valid() bool
}

type artifactoryClient struct {
	client     artifactory.ArtifactoryServicesManager
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

	// Note: do not reuse Vault request context here as this client is cached between requests.
	artifactoryServiceConfig, err := artconfig.NewConfigBuilder().
		SetServiceDetails(artifactoryDetails).
		SetHttpTimeout(config.ClientTimeout).
		// SetDryRun(false).
		SetContext(context.Background()).
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

func (ac *artifactoryClient) CreateToken(tokenReq TokenCreateEntry, role *RoleStorageEntry) (services.CreateTokenResponseData, error) {
	params := services.CreateTokenParams{
		Scope:     fmt.Sprintf("api:* member-of-groups:%s", groupName(role)),
		Username:  tokenUsername(role.Name),
		ExpiresIn: int(tokenReq.TTL.Seconds()),
	}

	return ac.client.CreateToken(params)
}
