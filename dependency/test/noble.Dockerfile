FROM ubuntu:noble

ENV DEBIAN_FRONTEND=noninteractive

# Runtime-focused package set for initial noble testing.
RUN apt-get update && \
  apt-get -y install \
    ca-certificates \
    curl \
    jq \
    openssl \
    libargon2-1 \
    libcurl4 \
    libedit2 \
    libenchant-2-2 \
    libffi8 \
    libfreetype6 \
    libgdbm6 \
    libgd3 \
    libgmp10 \
    libicu74 \
    libjpeg-turbo8 \
    libmaxminddb0 \
    libmemcached11 \
    libonig5 \
    libpng16-16 \
    libpq5 \
    libreadline8 \
    libsnmp40 \
    libsqlite3-0 \
    libssl3 \
    libtidy5deb1 \
    libwebp7 \
    libxml2 \
    libxslt1.1 \
    libyaml-0-2 \
    libzip4 \
    libfbclient2 \
    unixodbc \
    unzip \
    wget

COPY entrypoint.sh /entrypoint.sh
COPY fixtures /fixtures

ENTRYPOINT ["/entrypoint.sh"]
