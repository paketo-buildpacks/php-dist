# PHP Cloud Native Buildpack

The Cloud Foundry PHP Buildpack is a Cloud Native Buildpack V3 that provides PHP binaries to applications.

This buildpack is designed to work in collaboration with other buildpacks which request contributions of PHP.

## Detection

The detection phase always passes and contributes nothing to the build plan, depending on other buildpacks to request contributions.

## Build

If the build plan contains

- `php-binary`
  - Contributes PHP to a layer marked `build` and `cache` with all commands on `$PATH`
  - If `buildpack.yml` contains `php.verison`, configures a specific version.  This value must _exactly_ match a version available in the buildpack so typically it would configured to a wildcard such as `7.2.*`.
  - Contributes `$PHPRC` configured to the build layer
  - Contributes `$PHP_INI_SCAN_DIR` configured to the build layer
  - If `metadata.build = true`
    - Marks layer as `build` and `cache`
  - If `metadata.launch = true`
    - Marks layer as `launch`

## To Package

To package this buildpack for consumption:

```bash
$ ./scripts/package.sh
```

This builds the buildpack's Go source using GOOS=linux by default. You can supply another value as the first argument to package.sh.

## License
This buildpack is released under version 2.0 of the [Apache License][a].

[a]: http://www.apache.org/licenses/LICENSE-2.0