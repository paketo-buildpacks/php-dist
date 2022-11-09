FROM ubuntu:bionic

ENV DEBIAN_FRONTEND noninteractive

# See this list as the source for the apt packages installed
# https://github.com/paketo-buildpacks/dep-server/blob/3eb4dacd4be8ccca16bdf804c7308054563d98ba/.github/workflows/php-test-upload-metadata.yml#L19

RUN apt-get update && \
  apt-get -y install \
    jq \
    libargon2-0 \
    libcurl4 \
    libedit2 \
    libgd3 \
    libmagickwand-6.q16-3 \
    libonig4 \
    libpq5 \
    libxml2 \
    libxslt1-dev \
    libyaml-0-2

COPY entrypoint.sh /entrypoint.sh
COPY fixtures /fixtures

ENTRYPOINT ["/entrypoint.sh"]
