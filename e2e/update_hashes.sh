#!/usr/bin/env bash

set -e

cd "$(dirname "$0")"

# Source shared functions
source ./functions.sh

echo "Cleaning and regenerating files..."
git clean -dfx . > /dev/null 2>&1

echo "Running go generate..."
go generate "${GO_GENERATE_ARGS[@]}"

echo "Setting consistent modtimes..."
set_consistent_modtimes

echo "Computing hashes..."
compute_hashes "${GENERATED_HASHES_FILE}"

echo "Done! Updated ${GENERATED_HASHES_FILE}"
