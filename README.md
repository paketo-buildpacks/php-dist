# PHP Distribution Cloud Native Buildpack

The PHP Distribution CNB provides the PHP binary distribution. The buildpack
installs the PHP binary distribution onto the $PATH which makes it available
for subsequent buildpacks. These buildpacks can then use that distribution to
run PHP tooling. The PHP Web CNB is an example of a buildpack that utilizes the
PHP binary.

## Integration

The PHP Distribution CNB provides php as a dependency. Downstream buildpacks,
like [PHP Composer CNB](https://github.com/paketo-buildpacks/php-composer) can
require the php dependency by generating a [Build Plan
TOML](https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]

  # The name of the PHP dependency is "php". This value is considered
  # part of the public API for the buildpack and will not change without a plan
  # for deprecation.
  name = "php"

  # The PHP buildpack supports some non-required metadata options.
  [requires.metadata]

    # Setting the build flag to true will ensure that the PHP
    # depdendency is available on the $PATH for subsequent buildpacks during
    # their build phase. If you are writing a buildpack that needs to run PHP
    # during its build process, this flag should be set to true.
    build = true

    # Setting the launch flag to true will ensure that the PHP
    # dependency is available on the $PATH for the running application. If you are
    # writing an application that needs to run PHP at runtime, this flag should
    # be set to true.
    launch = true

    # Optional. If not provided, the buildpack will provide the default version from buildpack.toml.
    # To request a specific version, you can specify a semver constraint such as "8.*", "8.0.*",
    # or even "8.0.4".
    version = "8.0.4"
```

## Usage

To package this buildpack for consumption:

```bash
$ ./scripts/package.sh
```

This builds the buildpack's Go source using GOOS=linux by default. You can supply another value as the first argument to package.sh.

## License
This buildpack is released under version 2.0 of the [Apache License][a].

[a]: http://www.apache.org/licenses/LICENSE-2.0

## Run Tests

To run all unit tests, run:
```
./scripts/unit.sh
```

To run all integration tests, run:
```
./scripts/integration.sh
```

## Buildpack Configurations

### PHP Version
Specifying the PHP Dist version through buildpack.yml configuration is deprecated.

To migrate from using `buildpack.yml` please set the `$BP_PHP_VERSION`
environment variable at build time either directly (ex. `pack build my-app
--env BP_PHP_VERSION=7.3.*`) or through a [`project.toml`
file](https://github.com/buildpacks/spec/blob/main/extensions/project-descriptor.md)

```shell
# this allows you to specify a version constraint for the `php` depdendency
# any valid semver constaints (e.g. 7.*) are also acceptable
$BP_PHP_VERSION="7.3.*"
```
### PHP library directory
The PHP library directory is available to PHP via an include path in the PHP
configuration. By default it is set to `/workspace/lib` and can be overriden by
setting the `BP_PHP_LIB_DIR` environment variable at build-time.
```shell
$BP_PHP_LIB_DIR="some-directory"
```

### Provide custom `.ini` files
Custom `.ini` files can be provided from users to amend the default `php.ini`
file. This can be done by placing an `ini`-type configuration file inside
`<application directory>/.php.ini.d/`. Its path will be made available via the
`PHP_INI_SCAN_DIR`.

## Debug Logs
For extra debug logs from the image build process, set the `$BP_LOG_LEVEL`
environment variable to `DEBUG` at build-time (ex. `pack build my-app --env
BP_LOG_LEVEL=DEBUG` or through a  [`project.toml`
file](https://github.com/buildpacks/spec/blob/main/extensions/project-descriptor.md).
