#!/bin/bash

# Adapted from 

# script to normalize running make commands in CI. The essential difference in CI versus a local laptop
# is that the CI build runs in a container without access to the host network. So what we do is to run
# the build in a separate docker container which:
#
# * uses the host network
# * shares the common volume that is mounted into the current container
# * sets up the same environment variables and current directory of the current process
# * runs the "real" CI command
#
# For this to work, the default directory under which creds-helper writes creds needs to be somewhere
# in the shared volume so that the container we run can find it. The secondary container doesn't have
# access to creds-helper because the binary is an on-the-fly mount for the build container. (I suppose it is
# possible to inspect the container definition, extract this mount and re-apply it to the secondary container
# but I haven't done this)

# This is effectively an alternative to using GitLab Services - https://docs.gitlab.com/ee/ci/services/

set -euxo pipefail

entrypoint_script="$1"

# this section ensures that credentials are written under $(pwd) such that it is part of the
# shared volume and can be seen by the container we run.
mkdir -p tmp
export CREDS_SECRET_ROOT=$(pwd)/tmp


# save vars into a file (include DOCKER_CONFIG etc.) to be able to replay
# save only the var names and not the values which cause problems when the values have spaces.
# according to the docker docs just supplying a name causes the value to be picked up from the
# current environment.
set +x
env | grep  '[A-Za-z0-9_-][A-Za-z0-9_-]*=' | grep -v '^_' |  sort | awk -F= '{ print $1}' > tmp/vars.txt
set -x

# assume there is one running container and get its pid
pid=$(docker ps -q)

set +u
if [ "$CI_DEBUG_TRACE" = true ]; then
  # display current container mounts
  docker inspect ${pid} | jq '.[].Mounts' 
fi
set -u

# extract the image we are running from inspect output
image=$(docker inspect ${pid} | jq -r '.[0].Image')

# run it under the host network with the real CD entrypoint script
docker run --rm \
  --net=host \
  --env-file tmp/vars.txt \
  --volumes-from=${pid} \
  -w $(pwd) \
  --entrypoint $(pwd)/scripts/${entrypoint_script} \
  ${image}


