#!/bin/bash

set -eu

export PATH=$(pwd)/.tools:$PATH

DIR=$(dirname "$0")

. $DIR/wait_for.sh

pushd $DIR >&2
docker-compose version >&2
docker-compose up -d >&2

wait_for vault >&2
wait_for artifactory >&2


ARTIFACTORY_URL="http://localhost:8081/artifactory"

# install license key for Artifactory Pro (needed for creating groups/permission targets through API)
: "${ARTIFACTORY_LICENSE_KEY:?unset}"
payload=$(jq -n --arg lk "$ARTIFACTORY_LICENSE_KEY" '{licenseKey: $lk}')
curl -XPOST -H 'Content-type: application/json' -uadmin:password "${ARTIFACTORY_URL}/api/system/licenses" -d "$payload" >&2


# create a new admin user for UI use
password=$(openssl rand -base64 8)
payload=$(jq -n --arg pw "$password" '{userName: "dev", email: "dev@dev.net", password: $pw, admin: true}')
curl -XPOST -H 'Content-type: application/json' -uadmin:password "${ARTIFACTORY_URL}/api/security/users/dev" -d "$payload" >&2

# # change admin password
# username=admin
# password=$(openssl rand -base64 8)
# payload=$(jq -n --arg pw "$password" '{userName: "admin", oldPassword: "password", newPassword1: $pw, newPassword2: $pw}')
# curl -XPOST -H 'Content-type: application/json' -uadmin:password "${ARTIFACTORY_URL}/api/security/users/authorization/changePassword" -d "$payload" >&2

# eval output for local use
echo export ARTIFACTORY_USER="dev"\;
echo export ARTIFACTORY_PASSWORD=\"$password\"\;
echo export ARTIFACTORY_URL=\"$ARTIFACTORY_URL\"\;
echo export ARTIFACTORY_BEARER_TOKEN=$(curl -s -uadmin:$password -XPOST 'http://localhost:8081/artifactory/api/security/token' -d 'username=admin' -d 'expires_in=0' -d 'scope=member-of-groups:*' | jq -r .access_token)
echo export VAULT_ADDR="http://localhost:8200"\;


export VAULT_ADDR='http://localhost:8200'
export ARTIFACTORY_USER="dev";
export ARTIFACTORY_PASSWORD="$password"
# export ARTIFACTORY_BEARER_TOKEN=$(curl -s -u"${ARTIFACTORY_USER}:${ARTIFACTORY_PASSWORD}" -XPOST "${ARTIFACTORY_URL}/api/security/token" -d "username=$ARTIFACTORY_USER" -d 'expires_in=0' -d 'scope=member-of-groups:*' | jq -r .access_token)
export ARTIFACTORY_BEARER_TOKEN=$(curl -s -uadmin:$password -XPOST 'http://localhost:8081/artifactory/api/security/token' -d 'username=admin' -d 'expires_in=0' -d 'scope=member-of-groups:*' | jq -r .access_token)
# export ARTIFACTORY_BEARER_TOKEN=$(docker-compose exec artifactory cat var/etc/artifactory/security/access/access.admin.token)
export ARTIFACTORY_URL='http://artifactory:8081/artifactory'

"$DIR/setup_dev_artifactory.sh"
"$DIR/setup_dev_valt.sh"
