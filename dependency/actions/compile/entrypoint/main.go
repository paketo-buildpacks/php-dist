package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2/fs"

	"github.com/paketo-buildpacks/php-dist/actions/compile/entrypoint/internal"
)

func main() {
	var (
		version   string
		outputDir string
		target    string
	)

	flag.StringVar(&version, "version", "", "PHP dependency version")
	flag.StringVar(&outputDir, "outputDir", "", "Output directory for PHP dependency artifact")
	flag.StringVar(&target, "target", "", "Dependency target")
	flag.Parse()

	if version == "" {
		fail(errors.New(`missing required input "version"`))
	}

	if outputDir == "" {
		fail(errors.New(`missing required input "outputDir"`))
	}

	if target == "" {
		fail(errors.New(`missing required input "target"`))
	}

	dependencyDir, err := os.MkdirTemp("", "php-dist")
	if err != nil {
		log.Fatal(err)
	}

	err = installLibuv()
	if err != nil {
		log.Fatal(err)
	}

	err = runSymlinkCommands()
	if err != nil {
		log.Fatal(err)
	}

	err = installPHP(version, dependencyDir)
	if err != nil {
		log.Fatal(err)
	}

	err = internal.InstallExtensions(version, dependencyDir)
	if err != nil {
		log.Fatal(err)
	}

	err = setupTar(dependencyDir)
	if err != nil {
		log.Fatal(err)
	}

	err = createArtifact(version, target, outputDir, dependencyDir)
	if err != nil {
		log.Fatal(err)
	}
}

func setupTar(dependencyDir string) error {
	command := fmt.Sprintf(`
cp -a /usr/local/lib/x86_64-linux-gnu/librabbitmq.so* %[1]s/lib/;
cp -a /usr/local/lib/libhiredis.so* %[1]s/lib/;
cp -a /usr/lib/libc-client.so* %[1]s/lib/;
cp -a /usr/lib/libmcrypt.so* %[1]s/lib/;
cp -a /usr/lib/x86_64-linux-gnu/libaspell.so* %[1]s/lib/;
cp -a /usr/lib/x86_64-linux-gnu/libpspell.so* %[1]s/lib/;
cp -a /usr/lib/x86_64-linux-gnu/libmemcached.so* %[1]s/lib/;
cp -a /usr/local/lib/x86_64-linux-gnu/libcassandra.so* %[1]s/lib/;
cp -a /usr/local/lib/libuv.so* %[1]s/lib/;
cp -a /usr/lib/x86_64-linux-gnu/libargon2.so* %[1]s/lib/;
cp -a /usr/lib/librdkafka.so* %[1]s/lib/;
cp -a /usr/lib/x86_64-linux-gnu/libzip.so* %[1]s/lib/;
cp -a /usr/lib/x86_64-linux-gnu/libGeoIP.so* %[1]s/lib/;
cp -a /usr/lib/x86_64-linux-gnu/libgpgme.so* %[1]s/lib/;
cp -a /usr/lib/x86_64-linux-gnu/libassuan.so* %[1]s/lib/;
cp -a /usr/lib/x86_64-linux-gnu/libgpg-error.so* %[1]s/lib/;
cp -a /usr/lib/libtidy*.so* %[1]s/lib/;
cp -a /usr/lib/x86_64-linux-gnu/libtidy*.so* %[1]s/lib/;
cp -a /usr/lib/x86_64-linux-gnu/libenchant*.so* %[1]s/lib/;
cp -a /usr/lib/x86_64-linux-gnu/libfbclient.so* %[1]s/lib/;
cp -a /usr/lib/x86_64-linux-gnu/librecode.so* %[1]s/lib/;
cp -a /usr/lib/x86_64-linux-gnu/libtommath.so* %[1]s/lib/;
cp -a /usr/lib/x86_64-linux-gnu/libmaxminddb.so* %[1]s/lib/;
cp -a /usr/lib/x86_64-linux-gnu/libssh2.so* %[1]s/lib/
`, dependencyDir)

	err := internal.RunCommands(command, "", nil)
	if err != nil {
		return err
	}

	return nil
}

func createArtifact(version, target, outputDir, dependencyDir string) error {
	tempArchiveName := fmt.Sprintf("%s/temp.tgz", outputDir)

	log.Println("Building artifacts")
	err := internal.RunCommands(fmt.Sprintf("tar --create --gzip --verbose -f %s .", tempArchiveName), dependencyDir, nil)
	if err != nil {
		log.Println(fmt.Sprintf("Error building archive: %s", err))
		return err
	}

	sha256, err := fs.NewChecksumCalculator().Sum(tempArchiveName)
	if err != nil {
		return err
	}

	sha256 = sha256[0:64]
	outputTarballName := filepath.Join(outputDir, fmt.Sprintf("php_%s_linux_x64_%s_%s.tgz", version, target, sha256[0:8]))
	log.Printf("Exporting PHP tarball %s\n", filepath.Base(outputTarballName))
	err = fs.Move(tempArchiveName, outputTarballName)
	if err != nil {
		return err
	}

	log.Printf("Exporting PHP checksum file for %s\n", filepath.Base(outputTarballName))
	outputShaFileName := fmt.Sprintf("%s.checksum", outputTarballName)
	err = os.WriteFile(outputShaFileName, []byte(fmt.Sprintf("sha256:%s", sha256)), 0644)
	if err != nil {
		return err
	}

	return nil
}

func installLibuv() error {
	log.Println("Installing Libuv v1.12.0")

	err := internal.DownloadAndExtract(
		"libuv",
		"1.12.0",
		"http://dist.libuv.org/dist/v1.12.0/libuv-v1.12.0.tar.gz",
	)
	if err != nil {
		return err
	}

	dir := "libuv-1.12.0"
	command := "sh autogen.sh; ./configure; make; make install"

	err = internal.RunCommands(command, dir, nil)
	if err != nil {
		return err
	}

	return nil
}

func runSymlinkCommands() error {
	dir := ""
	command := `
ln -s /usr/include/x86_64-linux-gnu/curl /usr/local/include/curl;
ln -fs /usr/include/x86_64-linux-gnu/gmp.h /usr/include/gmp.h;
ln -fs /usr/lib/x86_64-linux-gnu/libldap.so /usr/lib/libldap.so;
ln -fs /usr/lib/x86_64-linux-gnu/libldap_r.so /usr/lib/libldap_r.so
`

	oracleExists, err := fs.Exists("/oracle")
	if err != nil {
		return err
	}

	if oracleExists {
		log.Println("Linking oracle lib files")
		command = command + ";ln -s /oracle/libclntsh.so.* /oracle/libclntsh.so"
	}

	err = internal.RunCommands(command, dir, nil)
	if err != nil {
		return err
	}

	return nil
}

func installPHP(version, dependencyDir string) error {
	log.Printf("Installing PHP v%s\n", version)

	err := internal.DownloadAndExtract(
		"php",
		version,
		fmt.Sprintf("https://github.com/php/web-php-distributions/raw/master/php-%s.tar.gz", version),
	)
	if err != nil {
		return err
	}

	command := fmt.Sprintf(`
./configure \
	--prefix=%s \
	--disable-static \
	--enable-shared \
	--enable-ftp=shared \
	--enable-sockets=shared \
	--enable-soap=shared \
	--enable-fileinfo=shared \
	--enable-bcmath \
	--enable-calendar \
	--enable-intl \
	--with-kerberos \
	--with-bz2=shared \
	--with-curl=shared \
	--enable-dba=shared \
	--with-password-argon2=/usr/lib/x86_64-linux-gnu \
	--with-cdb \
	--with-gdbm \
	--with-mysqli=shared \
	--enable-pdo=shared \
	--with-pdo-sqlite=shared,/usr \
	--with-pdo-mysql=shared,mysqlnd \
	--with-pdo-pgsql=shared \
	--with-pgsql=shared \
	--with-pspell=shared \
	--with-gettext=shared \
	--with-gmp=shared \
	--with-imap=shared \
	--with-imap-ssl=shared \
	--with-ldap=shared \
	--with-ldap-sasl \
	--with-zlib=shared \
	--with-libzip=/usr/local/lib \
	--with-xsl=shared \
	--with-snmp=shared \
	--enable-mbstring=shared \
	--enable-mbregex \
	--enable-exif=shared \
	--with-openssl=shared \
	--enable-fpm \
	--enable-pcntl=shared \
	--enable-sysvsem=shared \
	--enable-sysvshm=shared \
	--enable-sysvmsg=shared \
	--enable-shmop=shared;
make;
make install
`, dependencyDir)

	err = internal.RunCommands(command, fmt.Sprintf("php-%s", version), []string{"LIBS=-lz"})
	if err != nil {
		return err
	}

	return nil
}

func fail(err error) {
	log.Printf("Error: %s", err)
	os.Exit(1)
}
