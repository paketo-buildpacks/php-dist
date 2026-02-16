#!/usr/bin/env bash

set -euo pipefail

main() {
  local tarball_path expectedVersion
  tarball_path=""
  expectedVersion=""

  while [ "${#}" != 0 ]; do
    case "${1}" in
      --tarballPath)
        tarball_path="${2}"
        shift 2
        ;;

      --version)
        expectedVersion="${2}"
        shift 2
        ;;

      "")
        shift
        ;;

      *)
        echo "unknown argument \"${1}\""
        exit 1
    esac
  done

  if [[ "${tarball_path}" == "" ]]; then
    echo "--tarballPath is required"
    exit 1
  fi

  if [[ "${expectedVersion}" == "" ]]; then
    echo "--expectedVersion is required"
    exit 1
  fi

  echo "Outside image: tarball_path=${tarball_path}"
  echo "Outside image: expectedVersion=${expectedVersion}"

  local target
  if [[ $(basename "${tarball_path}") == *"jammy"* ]]; then
    target="jammy"
  else
    echo "compatible tests not found; skipping tests"
  fi

  echo "Running ${target} test..."
  docker build \
    --tag test \
    --file "${target}.Dockerfile" \
    .

  docker run \
    --rm \
    --volume "$(dirname -- "${tarball_path}"):/tarball_path" \
    test \
    --tarballPath "/tarball_path/$(basename "${tarball_path}")" \
    --expectedVersion "${expectedVersion}"
}

main "$@"
