#!/usr/bin/env bash

set -e

cd "$(dirname "$0")"

GENERATED_HASHES_FILE="generated_hashes.txt"
GO_GENERATE_ARGS="./..."

compare_hashes() {
  computed_hashes_file=$(mktemp)
  files=$(git ls-files . -o --ignored --exclude-standard)
  for file in ${files}; do
    hash=$(openssl dgst -r -md5 "${file}")
    # Detect stat version (GNU vs BSD)
    if stat --version &>/dev/null; then
      # GNU stat (Linux or from coreutils)
      size=$(stat -c "%s" "${file}")
    else
      # BSD stat (macOS)
      size=$(stat -f "%z" "${file}")
    fi

    echo "${size} ${hash}" >> "${computed_hashes_file}"
  done

  if ! diff -ru "${GENERATED_HASHES_FILE}" "${computed_hashes_file}"; then
    echo "Generated files are different. E2e failed."
    exit 1
  fi
}

# Use same modtimes for files
find . -type f -exec touch -m -t 202212311330 {} \;

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


