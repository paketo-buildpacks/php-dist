FROM ubuntu:resolute

ENV DEBIAN_FRONTEND=noninteractive

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
    libgdbm6t64 \
    libgd3 \
    libgmp10 \
    libicu78 \
    libjpeg-turbo8 \
    libmaxminddb0 \
    libmemcached11t64 \
    libonig5 \
    libpng16-16t64 \
    libpq5 \
    libreadline8t64 \
    libsasl2-2 \
    libsnmp40t64 \
    libsqlite3-0 \
    libssl3t64 \
    libtidy-dev \
    libwebp7 \
    libxml2-dev \
    libxslt1.1 \
    libyaml-0-2 \
    libzip-dev \
    libfbclient2 \
    unixodbc \
    unzip \
    wget

COPY entrypoint.sh /entrypoint.sh
COPY fixtures /fixtures

ENTRYPOINT ["/entrypoint.sh"]