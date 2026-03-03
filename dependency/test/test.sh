#!/usr/bin/env bash

set -euo pipefail

main() {
  local tarball_path expectedVersion os arch
  tarball_path=""
  expectedVersion=""
  os=""
  arch=""

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

      --os)
        os="${2}"
        shift 2
        ;;

      --arch)
        arch="${2}"
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
  elif [[ $(basename "${tarball_path}") == *"noble"* ]]; then
    target="noble"
  else
    echo "compatible tests not found; skipping tests"
    exit 0
  fi

  # When --os and --arch are provided, the --platform arg is passed to docker build and run commands.
  # This assumes the runner has qemu and buildkit set up, and that the docker daemon and cli experimental features are enabled.
  docker_platform_arg=""
  if [[ "${os}" != "" && "${arch}" != "" ]]; then
    docker_platform_arg="--platform ${os}/${arch}"
    echo "docker commands will be called with ${docker_platform_arg}"
  fi

  echo "Running ${target} test..."
  docker build \
    --tag test \
    --file "${target}.Dockerfile" \
    ${docker_platform_arg} \
    .

  docker run \
    --rm \
    --volume "$(dirname -- "${tarball_path}"):/tarball_path" \
    ${docker_platform_arg} \
    test \
    --tarballPath "/tarball_path/$(basename "${tarball_path}")" \
    --expectedVersion "${expectedVersion}"
}

main "$@"
