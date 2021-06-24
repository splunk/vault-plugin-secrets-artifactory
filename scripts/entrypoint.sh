#!/bin/bash

# This script serves as entrypoint to the reapplied ci job docker container.
# see test-wrapper.sh

set -uo pipefail

if [ $# -lt 1 ]; then
  echo "missing args: make target"
  exit 1
fi

target=${1}

echo "Current directory:"
pwd

set +u
if [ "$CI_DEBUG_TRACE" = true ]; then
  docker inspect $(docker ps -q)
fi
set -u

make $target
exit_code=$?

# any additional logic/debugging can go here

go-go vault --logout

exit ${exit_code}
