#!/usr/bin/env bash

set -euo pipefail
shopt -s inherit_errexit

parent_dir="$(cd "$(dirname "$0")" && pwd)"

major_minor_version=

verify_sha256sum() {
  tarballPath="${1}"
  dir=$(dirname "${tarballPath}")

  existing_sha256sum=$(cat "${tarballPath}.checksum")
  actual_sha256sum=$(sha256sum "${tarballPath}")
  actual_sha256sum="sha256:${actual_sha256sum:0:64}"

  if [[ "${actual_sha256sum}" != "${existing_sha256sum}" ]]; then
    echo "SHA256 ${actual_sha256sum} does not match expected SHA256 ${existing_sha256sum}"
    exit 1
  fi
}

extract_tarball() {
  rm -rf php
  mkdir php
  tar --extract --file "${1}" \
    --directory php
}

configure_php() {
  extension_dir="$(find "$PWD/php/lib/php/extensions" -name "no-debug-non-zts-*")"
  major_minor_version="$(./php/bin/php --version | head -1 | cut -d' ' -f2 | cut -d '.' -f1-2 | sed 's/\./-/')"

  sed \
    -i \
    "s|extension_dir=.*|extension_dir='${extension_dir}'|" \
    "php/bin/php-config"

  sed \
    "s|REPLACE_ME_EXTENSION_DIR|${extension_dir}|" \
    "${parent_dir}/fixtures/php.ini.template" \
    > "$PWD/php/etc/php.ini"

  <"${parent_dir}/fixtures/${major_minor_version}/extensions.json" \
    jq -r '.extensions // [] | .[]' | \
    sed 's/[",\,]//g' | \
    xargs -I {} echo "extension={}" >> "$PWD/php/etc/php.ini"

  <"${parent_dir}/fixtures/${major_minor_version}/extensions.json" \
    jq -r '.zend_extensions // [] | .[]' | \
    sed 's/[",\,]//g' | \
    xargs -I {} echo "zend_extension={}" >> "$PWD/php/etc/php.ini"

  export LD_LIBRARY_PATH="\$LD_LIBRARY_PATH:$PWD/php/lib"
}

check_version() {
  expected_version="${1}"
  actual_version="$(./php/bin/php --version | head -1 | cut -d' ' -f2)"
  if [[ "${actual_version}" != "${expected_version}" ]]; then
    echo "Version ${actual_version} does not match expected version ${expected_version}"
    exit 1
  fi
}

check_php_parsing() {
  set +e

  if ! ./php/bin/php -f "${parent_dir}/fixtures/hello.php" > output.html; then
    echo "Failed to run php"
    exit 1
  fi

  if ! diff "${parent_dir}/fixtures/hello.html" output.html > /dev/null; then
    echo "Actual output did not match expected output:"
    diff -u "${parent_dir}/fixtures/hello.html" output.html
    exit 1
  fi

  set -e
}

check_installed_modules() {
  set +e

  if ! ./php/bin/php -c "$PWD/php/etc/php.ini" -m > modules.txt; then
    echo "Failed to run php"
    exit 1
  fi

  if ! diff "${parent_dir}/fixtures/${major_minor_version}/expected-modules.txt" modules.txt > /dev/null; then
    echo "Could not find all expected modules. Diff:"
    diff -u "${parent_dir}/fixtures/${major_minor_version}/expected-modules.txt" modules.txt
    exit 1
  fi

  set -e
}

main() {
  local tarballPath expectedVersion
  tarballPath=""
  expectedVersion=""

  while [ "${#}" != 0 ]; do
    case "${1}" in
      --tarballPath)
        tarballPath="${2}"
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

  if [[ "${tarballPath}" == "" ]]; then
    echo "--tarballPath is required"
    exit 1
  fi

  if [[ "${expectedVersion}" == "" ]]; then
    echo "--expectedVersion is required"
    exit 1
  fi

  echo "Inside image: tarballPath=${tarballPath}"
  echo "Inside image: expectedVersion=${expectedVersion}"

  verify_sha256sum "${tarballPath}"
  extract_tarball "${tarballPath}"
  configure_php
  check_version "${expectedVersion}"
  check_php_parsing
  check_installed_modules

  echo "All tests passed!"
}

main "$@"
