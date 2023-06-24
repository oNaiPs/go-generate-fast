#!/usr/bin/env bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

cd "${DIR}"

pushd "${DIR}/.." > /dev/null
make build
popd  > /dev/null

"${DIR}/../bin/go-generate-fast" ./...