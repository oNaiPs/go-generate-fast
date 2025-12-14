#!/usr/bin/env bash

set -e

cd "$(dirname "$0")"

# Source shared functions
source ./functions.sh

compare_hashes() {
  computed_hashes_file=$(mktemp)
  compute_hashes "${computed_hashes_file}"

  if ! diff -ru "${GENERATED_HASHES_FILE}" "${computed_hashes_file}"; then
    echo "Generated files are different. E2e failed."
    exit 1
  fi
}

# Use same modtimes for files
set_consistent_modtimes

echo "- Running original go generate..."
git clean -dfx . > /dev/null 2>&1
go generate "${GO_GENERATE_ARGS[@]}"
compare_hashes

echo "- Running go-generate-fast with ONLY generation..."
git clean -dfx . > /dev/null 2>&1
GO_GENERATE_FAST_RECACHE=1 ../bin/go-generate-fast "${GO_GENERATE_ARGS[@]}"
compare_hashes

echo "- Running go-generate-fast with ONLY cache..."
git clean -dfx . > /dev/null 2>&1
GO_GENERATE_FAST_FORCE_USE_CACHE=1 ../bin/go-generate-fast "${GO_GENERATE_ARGS[@]}"
compare_hashes


