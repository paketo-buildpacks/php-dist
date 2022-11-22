FROM ubuntu:22.04

ENV DEBIAN_FRONTEND noninteractive

# See this list as the source for the apt packages installed
# https://github.com/paketo-buildpacks/dep-server/blob/3eb4dacd4be8ccca16bdf804c7308054563d98ba/.github/workflows/php-test-upload-metadata.yml#L19

# RUN apt-get update && \
#   apt-get -y install \
#     jq \
#     libargon2-0 \
#     libcurl4 \
#     libedit2 \
#     libgd3 \
#     libmagickwand-6.q16-6 \
#     libonig5 \
#     libpq5 \
#     libxml2 \
#     libxslt1-dev \
#     libsqlite0 \
#     libsqlite3-0 \
#     libyaml-0-2

RUN apt-get update && \
  apt-get -y install \
    ca-certificates \
    curl \
    dh-python \
    dnsutils \
    file \
    gir1.2-gdkpixbuf-2.0:amd64 \
    gir1.2-rsvg-2.0:amd64 \
    gnupg \
    gnupg1 \
    graphviz \
    gsfonts \
    gss-ntlmssp \
    imagemagick \
    imagemagick-6-common \
    jq \
    krb5-user \
    libaio1 \
    libarchive-extract-perl \
    libargon2-0 \
    libatm1 \
    libaudiofile1 \
    libavcodec58 \
    libbabeltrace1 \
    libblas3 \
    libc6 \
    libcurl4 \
    libdjvulibre-text \
    libdjvulibre21:amd64 \
    libdw1 \
    liberror-perl \
    libestr0 \
    libexif12 \
    libffi8 \
    libfl2 \
    libfribidi0 \
    libgcrypt20 \
    libgmp10 \
    libgmpxx4ldbl \
    libgnutls-openssl27 \
    libgnutls28-dev \
    libgnutls30 \
    libgnutlsxx28 \
    libgraphviz-dev \
    libharfbuzz-icu0 \
    libidn12 \
    libilmbase25:amd64 \
    libisl23:amd64 \
    libjson-glib-1.0-0 \
    libjsoncpp25:amd64 \
    liblapack3 \
    libldap-2.5-0 \
    liblockfile-bin \
    liblockfile1 \
    libmagic1 \
    libmariadb3 \
    libmodule-pluggable-perl \
    libmpc3:amd64 \
    libmpfr6:amd64 \
    libncurses5 \
    libnih-dbus1 \
    libnl-3-200:amd64 \
    libnl-genl-3-200:amd64 \
    libopenblas-base \
    libopenexr25:amd64 \
    liborc-0.4-0 \
    libp11-kit0 \
    libpam-cap \
    libpango1.0-0 \
    libpango1.0-dev \
    libpathplan4 \
    libpcre32-3 \
    libpq5 \
    libproxy1v5 \
    libpython3-stdlib:amd64 \
    libpython3.10 \
    libreadline8 \
    librhash0:amd64 \
    libsasl2-2 \
    libsasl2-modules \
    libsasl2-modules-gssapi-mit \
    libselinux1 \
    libsigc++-2.0-0v5:amd64 \
    libsigsegv2 \
    libsqlite0 \
    libsqlite3-0 \
    libsysfs2 \
    libtasn1-6 \
    libterm-ui-perl \
    libtiffxx5 \
    libtirpc-common:amd64 \
    libunwind8:amd64 \
    libustr-1.0-1 \
    libuv1:amd64 \
    libwmf0.2-7:amd64 \
    libwrap0:amd64 \
    libxapian30:amd64 \
    libxdot4 \
    libxslt1.1 \
    libyaml-0-2 \
    lockfile-progs \
    lsof \
    lzma \
    net-tools \
    ocaml-base-nox \
    openssh-client \
    openssl \
    psmisc \
    python3 \
    rsync \
    ruby \
    subversion \
    ubuntu-minimal \
    unixodbc \
    unzip \
    uuid \
    wget \
    zip

COPY entrypoint.sh /entrypoint.sh
COPY fixtures /fixtures

ENTRYPOINT ["/entrypoint.sh"]
