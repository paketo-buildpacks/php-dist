FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
      automake \
      build-essential \
      bundler \
      curl \
      cmake \
      firebird-dev \
      gcc \
      git \
      libargon2-dev \
      libaspell-dev \
      libbz2-dev \
      libc-client2007e-dev \
      libcurl4-openssl-dev \
      libedit-dev \
      libenchant-2-dev \
      libexpat1-dev \
      libgd-dev \
      libgdbm-dev \
      libgeoip-dev \
      libgmp-dev \
      libgpgme11-dev \
      libjpeg-dev \
      libkrb5-dev \
      libldap2-dev \
      libmagickcore-dev \
      libmagickwand-dev \
      libmaxminddb-dev \
      libmcrypt-dev \
      libmemcached-dev \
      libonig-dev \
      libpng-dev \
      libpq-dev \
      libpspell-dev \
      librecode-dev \
      libsasl2-dev \
      libsnmp-dev \
      libsqlite3-dev \
      libsqlite3-dev \
      libssh2-1-dev \
      libssl-dev \
      libtidy-dev \
      libtool \
      libwebp-dev \
      libxml2-dev \
      libxml2-dev \
      libxslt1-dev \
      libyaml-dev \
      libzip-dev \
      libzookeeper-mt-dev \
      make \
      pkg-config \
      ruby \
      ruby-dev \
      rubygems \
      snmp-mibs-downloader \
      sqlite3 \
      sudo \
      unixodbc-dev

# RUN apt-cache policy sqlite3

ADD ./extensions-manifests /tmp/extensions-manifests

ADD ./binary-builder /binary-builder
WORKDIR /binary-builder/cflinuxfs4
RUN bundle install


# add docker user to sudo-ers group
RUN useradd -m docker && echo "docker:docker" | chpasswd && adduser docker sudo

# make the user be docker, so it has sudo privileges
USER docker
# as currently written, the docker user has password docker and that needs to be manually entered to 'sudo'

ADD ./entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
