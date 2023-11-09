package internal

import (
	_ "embed"
	"fmt"
	"log"
	"path/filepath"
	"regexp"

	"github.com/paketo-buildpacks/packit/v2/fs"
)

type PeclExtensions struct {
	peclExts []PeclData
	phpDir   string
}

func (p PeclExtensions) Install() error {
	for _, peclExt := range p.peclExts {
		doInstall, err := shouldInstall(peclExt.Name)
		if err != nil {
			return err
		}

		if !doInstall {
			continue
		}

		log.Printf("Installing Pecl Extension, %s, v%s\n", peclExt.Name, peclExt.Version)

		workingDir := fmt.Sprintf("%s-%s", peclExt.Name, peclExt.Version)
		url := fmt.Sprintf("http://pecl.php.net/get/%s-%s.tgz", peclExt.Name, peclExt.Version)

		err = DownloadDependency(workingDir, url, peclExt.MD5)
		if err != nil {
			return err
		}

		if peclExt.Name == "maxminddb" {
			// commands for maxminddb must be run in subdirectory 'ext'
			workingDir = filepath.Join(workingDir, "ext")
		}

		options, err := getOptions(peclExt.Name)
		if err != nil {
			return err
		}

		command := fmt.Sprintf(
			"%[1]s/bin/phpize; ./configure --with-php-config=%[1]s/bin/php-config %[2]s; make; make install",
			p.phpDir,
			options,
		)

		err = RunCommands(command, workingDir, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func hasOracle() (bool, error) {
	oracleExists, err := fs.Exists("/oracle")
	if err != nil {
		return false, err
	}

	return oracleExists, nil
}

func oracleVersion() (string, error) {
	files, err := filepath.Glob("/oracle/")
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`libclntsh\.so\.`)
	for _, file := range files {
		if err != nil {
			return "", err
		}

		if !re.Match([]byte(file)) {
			continue
		}

		return string(re.ReplaceAll([]byte(file), []byte(""))), nil
	}

	return "", fmt.Errorf("not able to find oracle version")
}

func shouldInstall(ext string) (bool, error) {
	switch ext {
	case "oci8":
		return hasOracle()
	case "pdo_oci":
		return hasOracle()
	}

	return true, nil
}

type LocalExtensions struct {
	exts         []string
	version      string
	phpDir       string
	libsodiumDir string
}

func (l LocalExtensions) setupODBCTar(workingDir string) error {
	command := fmt.Sprintf(`
cp -a /usr/lib/x86_64-linux-gnu/libodbc.so* %[1]s/lib/;
cp -a /usr/lib/x86_64-linux-gnu/libodbcinst.so* %[1]s/lib/`, l.phpDir)

	err := RunCommands(command, "", nil)
	if err != nil {
		return err
	}

	return nil
}

func (l LocalExtensions) setupSodiumTar() error {
	command := fmt.Sprintf("cp -a /usr/local/lib/libsodium.so* %s/lib/", l.phpDir)

	err := RunCommands(command, "", nil)
	if err != nil {
		return err
	}

	return nil
}

func (l LocalExtensions) Install() error {
	for _, ext := range l.exts {
		doInstall, err := shouldInstall(ext)
		if err != nil {
			return err
		}

		if !doInstall {
			continue
		}

		log.Printf("Copying PHP extension files for %s\n", ext)

		envVars := []string{}

		workingDir := fmt.Sprintf("%s-%s", ext, l.version)

		err = fs.Copy(filepath.Join(fmt.Sprintf("php-%s", l.version), "ext", ext), workingDir)
		if err != nil {
			return err
		}

		if ext == "enchant" {
			log.Println("Applying patch for enchant")
			err = runEnchantPatch(workingDir)
			if err != nil {
				return err
			}
		} else if ext == "odbc" {
			log.Println("Applying patch for odbc")
			err = runODBCPatch(workingDir)
			if err != nil {
				return err
			}
		} else if ext == "sodium" {
			envVars = []string{
				fmt.Sprintf("LDFLAGS=-L%s/lib", l.libsodiumDir),
				fmt.Sprintf("PKG_CONFIG_PATH=%s/lib/pkgconfig", l.libsodiumDir),
			}
		}

		var options string
		if ext == "sodium" {
			options = fmt.Sprintf("--with-sodium=%s", l.libsodiumDir)
		} else {
			options, err = getOptions(ext)
			if err != nil {
				return err
			}
		}

		command := fmt.Sprintf(
			"%[1]s/bin/phpize; ./configure --with-php-config=%[1]s/bin/php-config %[2]s; make; make install",
			l.phpDir,
			options,
		)

		err = RunCommands(command, workingDir, envVars)
		if err != nil {
			return err
		}

		if ext == "odbc" || ext == "pdo_odbc" {
			err = l.setupODBCTar(workingDir)
			if err != nil {
				return err
			}
		} else if ext == "sodium" {
			err = l.setupSodiumTar()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func runODBCPatch(workingDir string) error {
	command := `echo 'AC_DEFUN([PHP_ALWAYS_SHARED],[])dnl' > temp.m4;
echo >> temp.m4;
cat config.m4 >> temp.m4;
mv temp.m4 config.m4;
`

	err := RunCommands(command, workingDir, nil)
	if err != nil {
		return err
	}

	return nil
}

func runEnchantPatch(workingDir string) error {
	command := `sed -i 's|#include "../spl/spl_exceptions.h"|#include <spl/spl_exceptions.h>|' enchant.c`
	err := RunCommands(command, workingDir, nil)
	if err != nil {
		return err
	}

	return nil
}

type TidewaysXhprof struct {
	data   ExtData
	phpDir string
}

func (t TidewaysXhprof) Install() error {
	log.Printf("Installing Tideways_Xhprof v%s\n", t.data.Version)

	workingDir := fmt.Sprintf("tideways_xhprof-%s", t.data.Version)
	url := fmt.Sprintf("https://github.com/tideways/php-xhprof-extension/archive/v%s.tar.gz", t.data.Version)
	err := DownloadDependency(workingDir, url, t.data.MD5)
	if err != nil {
		return err
	}

	command := fmt.Sprintf("%s/bin/phpize; ./configure --with-php-config=%s/bin/php-config; make; make install", t.phpDir, t.phpDir)

	err = RunCommands(command, workingDir, nil)
	if err != nil {
		return err
	}

	return nil
}

type Phpiredis struct {
	data       ExtData
	phpDir     string
	hiredisDir string
}

func (p Phpiredis) Install() error {
	log.Printf("Installing PHPIRedis, v%s, with hiredis dir, %s\n", p.data.Version, p.hiredisDir)

	workingDir := fmt.Sprintf("phpiredis-%s", p.data.Version)
	url := fmt.Sprintf("https://github.com/nrk/phpiredis/archive/v%s.tar.gz", p.data.Version)
	err := DownloadDependency(workingDir, url, p.data.MD5)
	if err != nil {
		return err
	}

	command := fmt.Sprintf(`
%s/bin/phpize;
./configure --with-php-config=%s/bin/php-config --enable-phpiredis --with-hiredis-dir=%s;
make;
make install
`, p.phpDir, p.phpDir, p.hiredisDir)

	err = RunCommands(command, workingDir, nil)
	if err != nil {
		return err
	}

	err = p.setupTar(workingDir)
	if err != nil {
		return err
	}

	return nil
}

func (p Phpiredis) setupTar(workingDir string) error {
	log.Println("Copying phpiredis to lib")

	command := fmt.Sprintf("cp -a modules/phpiredis.so* %s/lib", p.phpDir)

	err := RunCommands(command, workingDir, nil)
	if err != nil {
		return err
	}

	return nil
}

func getOptions(extName string) (string, error) {
	switch extName {
	case "redis":
		return "--enable-redis-igbinary --enable-redis-lzf --with-liblzf=no", nil
	case "oci8":
		return "--with-oci8=shared,instantclient,/oracle", nil
	case "memcached":
		return "--with-libmemcached-dir --enable-memcached-sasl --enable-memcached-msgpack --enable-memcached-igbinary --enable-memcached-json", nil
	case "odbc":
		return "--with-unixODBC=shared,/usr", nil
	case "gd":
		return "--with-external-gd", nil
	case "pdo_odbc":
		return "--with-pdo-odbc=unixODBC,/usr", nil
	case "pdo_oci":
		version, err := oracleVersion()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("--with-pdo-oci=shared,instantclient,/oracle,%s", version), nil
	default:
		return "", nil
	}
}
