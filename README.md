# PHP Dist Cloud Native Buildpack

The PHP Dist CNB provides the PHP binary distribution. The buildpack installs
the PHP binary distribution onto the $PATH which makes it available for
subsequent buildpacks. These buildpacks can then use that distribution to run
PHP tooling. The PHP Web CNB is an example of a buildpack that utilizes the PHP
binary.

## Integration

The PHP Dist CNB provides php as a dependency. Downstream buildpacks, like
[PHP Composer CNB](https://github.com/paketo-buildpacks/php-composer) can require the nginx
dependency by generating a [Build Plan
TOML](https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]

  # The name of the PHP dependency is "php". This value is considered
  # part of the public API for the buildpack and will not change without a plan
  # for deprecation.
  name = "php"

  # The version of the PHP dependency is not required. In the case it
  # is not specified, the buildpack will provide the default version, which can
  # be seen in the buildpack.toml file.
  # If you wish to request a specific version, the buildpack supports
  # specifying a semver constraint in the form of "7.*", "7.4.*", or even
  # "7.4.4".
  version = "7.4.4"

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
