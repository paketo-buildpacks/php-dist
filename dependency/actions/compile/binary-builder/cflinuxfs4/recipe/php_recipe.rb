# frozen_string_literal: true

require_relative 'php_common_recipes'

class PhpRecipe < BaseRecipe
  def configure_options
    [
      '--disable-static',
      '--enable-shared',
      '--enable-ftp=shared',
      '--enable-sockets=shared',
      '--enable-soap=shared',
      '--enable-fileinfo=shared',
      '--enable-bcmath',
      '--enable-calendar',
      '--enable-intl',
      '--with-kerberos',
      '--with-bz2=shared',
      '--with-curl=shared',
      '--enable-dba=shared',
      "--with-password-argon2=/usr/lib/#{multiarch_dir}",
      '--with-cdb',
      '--with-gdbm',
      '--with-mysqli=shared',
      '--enable-pdo=shared',
      '--with-pdo-sqlite=shared,/usr',
      '--with-pdo-mysql=shared,mysqlnd',
      '--with-pdo-pgsql=shared',
      '--with-pgsql=shared',
      '--with-pspell=shared',
      '--with-gettext=shared',
      '--with-gmp=shared',
      '--with-imap=shared',
      '--with-imap-ssl=shared',
      '--with-ldap=shared',
      '--with-ldap-sasl',
      '--with-zlib=shared',
      '--with-libzip=/usr/local/lib',
      '--with-xsl=shared',
      '--with-snmp=shared',
      '--enable-mbstring=shared',
      '--enable-mbregex',
      '--enable-exif=shared',
      '--with-openssl=shared',
      '--enable-fpm',
      '--enable-pcntl=shared',
      '--enable-sysvsem=shared',
      '--enable-sysvshm=shared',
      '--enable-sysvmsg=shared',
      '--enable-shmop=shared'
    ]
  end

  def multiarch_dir
    @multiarch_dir ||= `dpkg-architecture -qDEB_HOST_MULTIARCH`.strip
  end

  def url
    "https://github.com/php/web-php-distributions/raw/master/php-#{version}.tar.gz"
  end

  def archive_files
    ["#{port_path}/*"]
  end

  def archive_path_name
    'php'
  end

  def configure
    return if configured?

    md5_file = File.join(tmp_path, 'configure.md5')
    digest = Digest::MD5.hexdigest(computed_options.to_s)
    File.open(md5_file, 'w') { |f| f.write digest }

    # LIBS=-lz enables using zlib when configuring
    execute('configure', ['bash', '-c', "LIBS=-lz ./configure #{computed_options.join ' '}"])
  end

  def major_version
    @major_version ||= version.match(/^(\d+\.\d+)/)[1]
  end

  def zts_path
    Dir["#{path}/lib/php/extensions/no-debug-non-zts-*"].first
  end

  def setup_tar
    lib_dir = "/usr/lib/#{multiarch_dir}"
    argon_dir = "/usr/lib/#{multiarch_dir}"
    local_lib_dir = "/usr/local/lib/#{multiarch_dir}"

    system <<-EOF
      cp -a -v #{local_lib_dir}/librabbitmq.so* #{path}/lib/
      cp -a -v #{@hiredis_path}/lib/libhiredis.so* #{path}/lib/
      cp -a -v /usr/lib/libc-client.so* #{path}/lib/
      cp -a -v /usr/lib/libmcrypt.so* #{path}/lib
      cp -a -v #{lib_dir}/libmcrypt.so* #{path}/lib
      cp -a -v #{lib_dir}/libaspell.so* #{path}/lib
      cp -a -v #{lib_dir}/libpspell.so* #{path}/lib
      cp -a -v #{lib_dir}/libmemcached.so* #{path}/lib/
      cp -a -v #{local_lib_dir}/libcassandra.so* #{path}/lib
      cp -a -v /usr/local/lib/libuv.so* #{path}/lib
      cp -a -v #{argon_dir}/libargon2.so* #{path}/lib
      cp -a -v /usr/lib/librdkafka.so* #{path}/lib
      cp -a -v #{lib_dir}/libzip.so* #{path}/lib/
      cp -a -v #{lib_dir}/libGeoIP.so* #{path}/lib/
      cp -a -v #{lib_dir}/libgpgme.so* #{path}/lib/
      cp -a -v #{lib_dir}/libassuan.so* #{path}/lib/
      cp -a -v #{lib_dir}/libgpg-error.so* #{path}/lib/
      cp -a -v /usr/lib/libtidy*.so* #{path}/lib/
      cp -a -v #{lib_dir}/libtidy*.so* #{path}/lib/
      cp -a -v #{lib_dir}/libenchant*.so* #{path}/lib/
      cp -a -v #{lib_dir}/libfbclient.so* #{path}/lib/
      cp -a -v #{lib_dir}/librecode.so* #{path}/lib/
      cp -a -v #{lib_dir}/libtommath.so* #{path}/lib/
      cp -a -v #{lib_dir}/libmaxminddb.so* #{path}/lib/
      cp -a -v #{lib_dir}/libssh2.so* #{path}/lib/
    EOF

    system "cp #{@ioncube_path}/ioncube/ioncube_loader_lin_#{major_version}.so #{zts_path}/ioncube.so" if IonCubeRecipe.build_ioncube?(version)

    system <<-EOF
      # Remove unused files
      rm "#{path}/etc/php-fpm.conf.default"
      rm "#{path}/bin/php-cgi"
      find "#{path}/lib/php/extensions" -name "*.a" -type f -delete
    EOF
  end
end
