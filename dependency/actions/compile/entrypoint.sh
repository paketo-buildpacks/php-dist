#!/usr/bin/env bash

set -eu
set -o pipefail
shopt -s inherit_errexit

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

function main() {
  local version output_dir target bundler_dir

  while [ "${#}" != 0 ]; do
    case "${1}" in
      --version)
        version="${2}"
        shift 2
        ;;

      --outputDir)
        output_dir="${2}"
        shift 2
        ;;

      --target)
        target="${2}"
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

  if [[ -z "${version:-}" ]]; then
    echo "--version is required"
    exit 1
  fi

  if [[ -z "${output_dir:-}" ]]; then
    echo "--outputDir is required"
    exit 1
  fi

  if [[ -z "${target:-}" ]]; then
    echo "--target is required"
    exit 1
  fi

  echo "Validating PHP version + target combination"
  if [[ ${version} == "7.4"* || ${version} == "8.0"* ]]; then
    if grep -q "Jammy" "/etc/os-release"; then
    echo "Cannot build PHP ${version} on Jammy"
    exit 1
    fi
  fi

  echo "Downloading source from upstream"
  local upstream="/tmp/upstream.tgz"
  curl "https://github.com/php/web-php-distributions/raw/master/php-${version}.tar.gz" \
    --silent \
    --fail \
    --show-error \
    --output "${upstream}"

  echo "Determining extensions file"
  local extensions_file

  if [[ ${version} == "7.4"* ]]; then
    extensions_file="extensions-7.4.yml"
  elif [[ ${version} == "8.0"* ]]; then
    extensions_file="extensions-8.0.yml"
  elif [[ ${version} == "8.1"* ]]; then
    extensions_file="extensions-8.1.yml"
  else
    echo "No extensions file found for PHP version ${version}"
    exit 1
  fi

  echo "Calculating upstream checksum"
  upstream_sha="$(sha256sum "${upstream}" | cut -d " " -f 1 )"
  echo "${upstream_sha}"

  echo "Compiling PHP with extensions from ${extensions_file}"
  echo "${PASSWORD}" | sudo -S /usr/bin/ruby -x /usr/bin/bundler exec /usr/bin/ruby ./bin/binary-builder.rb \
    --name php \
    --version "${version}" \
    --sha256 "${upstream_sha}" \
    --php-extensions-file "/tmp/extensions-manifests/${extensions_file}"

  echo "Packaging php dependency"
  cp ./php-"${version}"*.tgz "${output_dir}/php-${target}-${version}.tgz"
  echo "Wrote php-${target}-${version}.tgz to ${output_dir}"

  # extract SHA256 from sha256sum output in the formm
  # someshaabcdefgh /path/to/file
  sha="$(sha256sum "${output_dir}/php-${target}-${version}.tgz" | cut -d " " -f 1 )"
  echo "sha256:${sha}" > "${output_dir}/php-${target}-${version}.tgz.checksum"
  echo "Wrote php-${target}-${version}.tgz.checksum to ${output_dir}"
}

main "${@:-}"
