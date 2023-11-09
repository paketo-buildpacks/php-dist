FROM paketobuildpacks/build-bionic-full

ENV DEBIAN_FRONTEND noninteractive

ARG cnb_uid=0
ARG cnb_gid=0

USER ${cnb_uid}:${cnb_gid}

RUN apt-get update && \
  apt-get -y install \
  firebird-dev \
  gcc \
  libaspell-dev \
  libbz2-dev \
  libc-client2007e-dev \
  libedit-dev \
  libenchant-dev \
  libexpat1-dev \
  libgdbm-dev \
  libgeoip-dev \
  libgpgme11-dev \
  libjpeg-dev \
  libmagickcore-dev \
  libmaxminddb-dev \
  libmcrypt-dev \
  libmemcached-dev \
  libonig-dev \
  libpng-dev \
  libpspell-dev \
  librecode-dev \
  libsnmp-dev \
  libssh2-1-dev \
  libtidy-dev \
  libwebp-dev \
  libxml2-dev \
  libzip-dev \
  libzookeeper-mt-dev \
  make \
  pkg-config \
  snmp-mibs-downloader \
  software-properties-common \
  sqlite3 \
  sudo

RUN add-apt-repository ppa:longsleep/golang-backports
RUN apt-get -y install golang-go

COPY entrypoint /tmp/entrypoint
RUN cd /tmp/entrypoint && go build -o /entrypoint .

ENTRYPOINT ["/entrypoint"]
