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
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

const (
	pluginPrefix         = "vault-plugin"
	tokenUsernamePrefix  = "auto-vault-plugin"
	tokenUsernameMaxLen  = 58
	tokenUsernameHashLen = 8
)

func groupName(roleEntry *RoleStorageEntry) string {
	return fmt.Sprintf("%s.%s", pluginPrefix, roleEntry.RoleID)
}

func permissionTargetName(roleName string, index int) string {
	return fmt.Sprintf("%s.pt%d.%s", pluginPrefix, index, roleName)
}

func tokenUsername(roleName string) string {
	fullUsername := fmt.Sprintf("%s.%s", tokenUsernamePrefix, roleName)
	tokenUser := fullUsername
	if len(fullUsername) > tokenUsernameMaxLen {
		truncIndex := tokenUsernameMaxLen - tokenUsernameHashLen
		h := fmt.Sprintf("%x", sha256.Sum256([]byte(fullUsername[truncIndex:])))
		tokenUser = fullUsername[:truncIndex] + h[:tokenUsernameHashLen]
	}
	return tokenUser
}

// appendTrailingSlash appends trailing slash if url doesn't end with slash.
// artifactory client assumes URL ends with '/'
func appendTrailingSlash(url string) string {
	if !strings.HasSuffix(url, "/") {
		return fmt.Sprintf("%s/", url)
	}
	return url
}

func convertPermissionTarget(fromPt *PermissionTarget, toPt *services.PermissionTargetParams, groupName, ptName string) {

	if fromPt.Repo != nil {
		groupRepo := map[string][]string{
			groupName: fromPt.Repo.Operations,
		}
		p := &services.PermissionTargetSection{
			IncludePatterns: fromPt.Repo.IncludePatterns,
			ExcludePatterns: fromPt.Repo.ExcludePatterns,
			Repositories:    fromPt.Repo.Repositories,
			Actions:         &services.Actions{Groups: groupRepo},
		}
		toPt.Repo = p
	}

	if fromPt.Build != nil {

		groupBuild := map[string][]string{
			groupName: fromPt.Build.Operations,
		}
		p := &services.PermissionTargetSection{
			IncludePatterns: fromPt.Build.IncludePatterns,
			ExcludePatterns: fromPt.Build.ExcludePatterns,
			Repositories:    fromPt.Build.Repositories,
			Actions:         &services.Actions{Groups: groupBuild},
		}
		toPt.Build = p
	}

	toPt.Name = ptName
}

func validateOperations(ops []string) error {
	var err *multierror.Error

	for _, op := range ops {
		switch op {
		case "read", "write", "annotate",
			"delete", "manage", "managedXrayMeta",
			"distribute":
			continue
		default:
			err = multierror.Append(err, fmt.Errorf("operation '%s' is not allowed", op))
		}
	}

	return err.ErrorOrNil()
}

func getStringHash(ptsRaw string) string {
	ssum := sha256.Sum256([]byte(ptsRaw))
	return base64.StdEncoding.EncodeToString(ssum[:])
}
