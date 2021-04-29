#!/bin/bash

set -euo pipefail

ARTIFACTORY_URL="http://localhost:8081/artifactory/"
ARTIFACTORY_USER="admin";
ARTIFACTORY_PASSWORD="password"
ARTIFACTORY_BEARER_TOKEN=$(curl -s -u"${ARTIFACTORY_USER}:${ARTIFACTORY_PASSWORD}" -XPOST "${ARTIFACTORY_URL}api/security/token" -d "username=$ARTIFACTORY_USER" -d 'expires_in=0' -d 'scope=member-of-groups:*' | jq -r .access_token)

setup_artifactory() {
  auth_header="Authorization: Bearer $ARTIFACTORY_BEARER_TOKEN"
  content_header='Content-Type: application/json'

  # install license key for Artifactory Pro (required to enable all API endpoints)
  installed=$(curl -sSH "$auth_header" "${ARTIFACTORY_URL}api/system/licenses")

  if [ -n "$(echo "$installed" | jq -r .licensedTo)" ]; then
    echo
    echo "License key already installed:"
    echo "$installed" | jq
  else
    echo "Installing Artifactory license key..."
    payload=$(jq -n --arg lk "$ARTIFACTORY_LICENSE_KEY" '{licenseKey: $lk}')
    echo "License key install response status:"
    curl -sS -XPOST -H "$auth_header" -H "$content_header" "${ARTIFACTORY_URL}api/system/licenses" -d "$payload" | jq .status
  fi


  # create some local repos to use with sample permission targets
  payload=$(jq -n '{rclass: "local"}')
  for repo in "docker-test-west-local" "docker-test-east-local" "npm-test-west-local" "npm-test-east-local"; do
    curl -sS -XPUT -H "$auth_header" -H "$content_header" "${ARTIFACTORY_URL}api/repositories/${repo}" -d "$payload" || true
    echo
  done

  # create an apiKey for admin user if it doesn't already exist
  ARTIFACTORY_API_KEY=$(curl -s -u"${ARTIFACTORY_USER}:${ARTIFACTORY_PASSWORD}" "${ARTIFACTORY_URL}api/security/apiKey" | jq -r .apiKey)
  if [ -z "$ARTIFACTORY_API_KEY" ] || [ "$ARTIFACTORY_API_KEY" == "null" ]; then
    echo "Creating new api key..."
    ARTIFACTORY_API_KEY=$(curl -s -u"${ARTIFACTORY_USER}:${ARTIFACTORY_PASSWORD}" -XPOST "${ARTIFACTORY_URL}api/security/apiKey" | jq -r .apiKey)
  fi

  # # create a new admin user for UI use
  # password=$(openssl rand -base64 8)
  # payload=$(jq -n --arg pw "$password" '{userName: "dev", email: "dev@dev.net", password: $pw, admin: true}')
  # curl -XPOST -H 'Content-type: application/json' -uadmin:password "${ARTIFACTORY_URL}api/security/users/dev" -d "$payload" >&2

  # # change admin password
  # username=admin
  # password=$(openssl rand -base64 8)
  # payload=$(jq -n --arg pw "$password" '{userName: "admin", oldPassword: "password", newPassword1: $pw, newPassword2: $pw}')
  # curl -XPOST -H 'Content-type: application/json' -uadmin:password "${ARTIFACTORY_URL}api/security/users/authorization/changePassword" -d "$payload" >&2
}

setup_artifactory >&2

# eval output for local use
echo export ARTIFACTORY_USER=\"$ARTIFACTORY_USER\"\;
echo export ARTIFACTORY_PASSWORD=\"$ARTIFACTORY_PASSWORD\"\;
echo export ARTIFACTORY_URL=\"$ARTIFACTORY_URL\"\;
echo export ARTIFACTORY_BEARER_TOKEN=\"$ARTIFACTORY_BEARER_TOKEN\"\;
echo export ARTIFACTORY_API_KEY=\"$ARTIFACTORY_API_KEY\"\;
