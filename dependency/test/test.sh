#!/usr/bin/env bash

set -euo pipefail
#shopt -s inherit_errexit

parent_dir="$(cd "$(dirname "$0")" && pwd)"

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

      --expectedVersion)
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

  if [[ $(basename "${tarball_path}") == *"bionic"* ]]; then
    echo "Running bionic test..."
    docker build \
      --tag test \
      --file bionic.Dockerfile \
      .

    docker run \
      --rm \
      --volume "$(dirname -- "${tarball_path}"):/tarball_path" \
      test \
      --tarballPath "/tarball_path/$(basename "${tarball_path}")" \
      --expectedVersion "${expectedVersion}"
  else
    echo "bionic not found - skipping tests"
  fi
}

main "$@"
