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
  if [[ ${version} == "8.0"* ]]; then
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

  if [[ ${version} == "8.0"* ]]; then
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
  /usr/bin/ruby -x /usr/bin/bundler exec /usr/bin/ruby ./bin/binary-builder.rb \
    --name php \
    --version "${version}" \
    --sha256 "${upstream_sha}" \
    --php-extensions-file "/tmp/extensions-manifests/${extensions_file}"

  archive="${output_dir}/php-${target}-${version}.tgz"
  cp ./php-"${version}"*.tgz "${archive}"
  strip_dir="${output_dir}/strip_dir"
  rm -rf "${strip_dir}"
  mkdir "${strip_dir}"
  tar -C "${strip_dir}" --transform s:^\./:: --strip-components 1 -xf "${archive}"
  tar -C "${strip_dir}" -czf "${archive}" .
  rm -rf strip_dir

  echo "Stripped top-level directory from tar"

  SHA256=$(sha256sum "${archive}")
  SHA256="${SHA256:0:64}"

  OUTPUT_TARBALL_NAME="php_${version}_linux_x64_${target}_${SHA256:0:8}.tgz"
  OUTPUT_SHAFILE_NAME="php_${version}_linux_x64_${target}_${SHA256:0:8}.tgz.checksum"

  echo "Building tarball ${OUTPUT_TARBALL_NAME}"

  mv "${output_dir}/php-${target}-${version}.tgz" "${output_dir}/${OUTPUT_TARBALL_NAME}"

  echo "Creating checksum file for ${OUTPUT_TARBALL_NAME}"
  echo "sha256:${SHA256}" > "${output_dir}/${OUTPUT_SHAFILE_NAME}"
}

main "${@:-}"
