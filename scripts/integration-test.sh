#!/bin/bash

# This script serves as entrypoint to the reapplied job docker container.
# see test-wrapper.sh

set -euxo pipefail

set +e
make integration-test
exit_code=$?

# any additional logic/debugging can go here

go-go vault --logout

exit ${exit_code}
