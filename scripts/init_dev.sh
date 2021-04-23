#!/bin/bash

set -euo pipefail

export PATH=$(pwd)/.tools:$PATH

DIR=$(dirname "$0")

. $DIR/wait_for.sh

pushd $DIR &>/dev/null
docker-compose version >&2

set +u
if [ -n "$CI" ]; then
  echo "Starting containers in net=host mode..."
  docker-compose -f docker-compose-ci.yaml up -d
else
  docker-compose up -d >&2
fi
set -u

wait_for vault >&2
wait_for artifactory >&2

export ARTIFACTORY_URL="http://localhost:8081/artifactory"
export ARTIFACTORY_USER="admin";
export ARTIFACTORY_PASSWORD="password"
export ARTIFACTORY_BEARER_TOKEN=$(curl -s -u"${ARTIFACTORY_USER}:${ARTIFACTORY_PASSWORD}" -XPOST "${ARTIFACTORY_URL}/api/security/token" -d "username=$ARTIFACTORY_USER" -d 'expires_in=0' -d 'scope=member-of-groups:*' | jq -r .access_token)
auth="Bearer $ARTIFACTORY_BEARER_TOKEN"

# install license key for Artifactory Pro (required to enable all API endpoints)
installed=$(curl -sSH "Authorization: $auth" "${ARTIFACTORY_URL}/api/system/licenses")

if [ -n "$(echo "$installed" | jq -r .licensedTo)" ]; then
  echo
  echo "License key already installed:" >&2
  echo "$installed" | jq >&2
else
  echo "Installing Artifactory license key..." >&2
  payload=$(jq -n --arg lk "$ARTIFACTORY_LICENSE_KEY" '{licenseKey: $lk}')
  echo "License key install response status:" >&2
  curl -sS -XPOST -H "Authorization: $auth" -H 'Content-type: application/json' "${ARTIFACTORY_URL}/api/system/licenses" -d "$payload" | jq .status >&2
fi

# # create a new admin user for UI use
# password=$(openssl rand -base64 8)
# payload=$(jq -n --arg pw "$password" '{userName: "dev", email: "dev@dev.net", password: $pw, admin: true}')
# curl -XPOST -H 'Content-type: application/json' -uadmin:password "${ARTIFACTORY_URL}/api/security/users/dev" -d "$payload" >&2

# # change admin password
# username=admin
# password=$(openssl rand -base64 8)
# payload=$(jq -n --arg pw "$password" '{userName: "admin", oldPassword: "password", newPassword1: $pw, newPassword2: $pw}')
# curl -XPOST -H 'Content-type: application/json' -uadmin:password "${ARTIFACTORY_URL}/api/security/users/authorization/changePassword" -d "$payload" >&2

export VAULT_ADDR="http://localhost:8200"
export VAULT_TOKEN=root

set +u
if [ -z "$CI" ]; then
  # eval output for local use
  echo -e "\n\033[1;33mCopy/paste or eval this script:\033[0m\n" >&2
  echo export ARTIFACTORY_USER=\"$ARTIFACTORY_USER\"\;
  echo export ARTIFACTORY_PASSWORD=\"$ARTIFACTORY_PASSWORD\"\;
  echo export ARTIFACTORY_URL=\"$ARTIFACTORY_URL\"\;
  echo export ARTIFACTORY_BEARER_TOKEN=\"$ARTIFACTORY_BEARER_TOKEN\"\;
  echo export VAULT_ADDR=\"$VAULT_ADDR\"\;
  echo export VAULT_TOKEN=\"$VAULT_TOKEN\"\;

  # containers using net="host" in CI will all communicate via localhost
  export ARTIFACTORY_URL='http://artifactory:8081/artifactory'
fi
set -u


popd &>/dev/null

echo "Configuring vault..." >&2
"$DIR/setup_dev_vault.sh" >&2

echo -e "\nExample usage to test plugin:" >&2
echo -e "\033[0;32mvault write artifactory/roles/role1 token_ttl=600 permission_targets=@scripts/sample_permission_targets.json\033[0m\n" >&2
