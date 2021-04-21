package artifactorysecrets

import (
	"net/http"

	v1 "github.com/atlassian/go-artifactory/v2/artifactory/v1"
)

type mockArtifactoryClient struct {
	// client *artifactory.Artifactory
	// context context.Context
}

var _ Client = &mockArtifactoryClient{}

func (ac *mockArtifactoryClient) CreateOrReplaceGroup(role *RoleStorageEntry) (*http.Response, error) {
	return nil, nil
}

func (ac *mockArtifactoryClient) DeleteGroup(role *RoleStorageEntry) (*string, *http.Response, error) {
	return nil, nil, nil
}
func (ac *mockArtifactoryClient) CreateOrUpdatePermissionTarget(role *RoleStorageEntry, pt *PermissionTarget) (*http.Response, error) {
	return nil, nil
}
func (ac *mockArtifactoryClient) DeletePermissionTarget(role *RoleStorageEntry, pt *PermissionTarget) (*http.Response, error) {
	return nil, nil
}
func (ac *mockArtifactoryClient) CreateToken(tokenReq TokenCreateEntry, role *RoleStorageEntry) (*v1.AccessToken, *http.Response, error) {
	return nil, nil, nil
}
