package artifactorysecrets

import (
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

type mockArtifactoryClient struct {
	// client *artifactory.Artifactory
	// context context.Context
}

var _ Client = &mockArtifactoryClient{}

func (ac *mockArtifactoryClient) CreateOrReplaceGroup(role *RoleStorageEntry) error {
	return nil
}

func (ac *mockArtifactoryClient) DeleteGroup(role *RoleStorageEntry) error {
	return nil
}
func (ac *mockArtifactoryClient) CreateOrUpdatePermissionTarget(role *RoleStorageEntry, pt *PermissionTarget, ptName string) error {
	return nil
}
func (ac *mockArtifactoryClient) DeletePermissionTarget(ptName string) error {
	return nil
}
func (ac *mockArtifactoryClient) CreateToken(tokenReq TokenCreateEntry, role *RoleStorageEntry) (services.CreateTokenResponseData, error) {
	return services.CreateTokenResponseData{}, nil
}
