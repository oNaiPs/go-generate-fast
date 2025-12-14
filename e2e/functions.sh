#!/usr/bin/env bash

# Shared functions and variables for e2e tests

GENERATED_HASHES_FILE="generated_hashes.txt"
GO_GENERATE_ARGS="./..."

set_consistent_modtimes() {
  find . -type f -exec touch -m -t 202212311330 {} \;
}

compute_hashes() {
  local output_file="${1}"
  > "${output_file}"

  files=$(git ls-files . -o --ignored --exclude-standard)
  for file in ${files}; do
    hash=$(openssl dgst -r -md5 "${file}")
    # Detect stat version (GNU vs BSD)
    if stat --version &>/dev/null 2>&1; then
      # GNU stat (Linux or from coreutils)
      size=$(stat -c "%s" "${file}")
    else
      # BSD stat (macOS)
      size=$(stat -f "%z" "${file}")
    fi

    echo "${size} ${hash}" >> "${output_file}"
  done
}
