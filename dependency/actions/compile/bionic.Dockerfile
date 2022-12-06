FROM ubuntu:18.04

RUN apt-get update && apt-get install -y \
      automake \
      build-essential \
      bundler \
      cmake \
      curl \
      firebird-dev \
      gcc \
      git \
      libargon2-0-dev \
      libaspell-dev \
      libbz2-dev \
      libc-client2007e-dev \
      libcurl4-openssl-dev \
      libedit-dev \
      libenchant-dev \
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
      rbenv \
      ruby \
      ruby-dev \
      rubygems \
      snmp-mibs-downloader \
      sqlite3 \
      sudo \
      unixodbc-dev

ADD ./extensions-manifests /tmp/extensions-manifests

ADD ./binary-builder /binary-builder
WORKDIR /binary-builder

RUN bundle install

ADD ./entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
