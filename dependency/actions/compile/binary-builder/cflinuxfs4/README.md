# Introduction

This tool provides a mechanism for building php binaries. This subdirectory is
a fork of its parent directory. This one uses Ruby 3 for compatibility with
running on Jammy Jellyfish.

# Usage

TODO: deal with "run as root" The scripts are **meant to be run as root** on a
Cloud Foundry [stack](https://docs.cloudfoundry.org/concepts/stacks.html).

# Building PHP

To build PHP, you also need to pass in a YAML file containing information about
the various PHP extensions to be built.

For an example of what these files look like, see:
[php7-base-extensions.yml](../extensions-manifests).

**TIP** If you are updating or building a specific PHP extension, remove
everything except that specific extension from your `--php-extensions-file`
file. This will decrease the build times & make it easier for you to test your
changes.

# Running the tests
                                                                                                                                                              │ 24       libgpgme11-dev \
```bash                                                                                                                                                       │ 25       libjpeg-dev \
bundle                                                                                                                                                        │ 26       libkrb5-dev \
bundle exec rspec                                                                                                                                             │ 27       libldap2-dev \
```

The integration test suite includes specs that test the functionality for
building [PHP with Oracle client libraries](./PHP-Oracle.md). These tests are
tagged `:run_oracle_php_tests` and require access to an S3 bucket containing
the Oracle client libraries. This is configured using the environment variables
`AWS_ACCESS_KEY` and `AWS_SECRET_ACCESS_KEY`

If you do not need to test this functionality, exclude the tag
`:run_oracle_php_tests` when you run `rspec`.
