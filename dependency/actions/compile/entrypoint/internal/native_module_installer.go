package internal

import (
	_ "embed"
	"fmt"
	"log"
	"path/filepath"
)

type Snmp struct {
	phpDir string
}

func (s Snmp) Install() error {
	log.Println("Installing SNMP")

	command := `
mkdir -p mibs;
cp -a /usr/lib/x86_64-linux-gnu/libnetsnmp.so* lib/;

cp -r /usr/share/snmp/mibs/* mibs;

cp /usr/bin/download-mibs bin;
cp /usr/bin/smistrip bin;
sed -i "s|^CONFDIR=/etc/snmp-mibs-downloader|CONFDIR=\$HOME/php/mibs/conf|" bin/download-mibs;
sed -i "s|^SMISTRIP=/usr/bin/smistrip|SMISTRIP=\$HOME/php/bin/smistrip|" bin/download-mibs;

cp -R /etc/snmp-mibs-downloader mibs/conf;
sed -i "s|^DIR=/usr/share/doc|DIR=\$HOME/php/mibs/originals|" mibs/conf/iana.conf;
sed -i "s|^DEST=iana|DEST=|" mibs/conf/iana.conf;
sed -i "s|^DIR=/usr/share/doc|DIR=\$HOME/php/mibs/originals|" mibs/conf/ianarfc.conf;
sed -i "s|^DEST=iana|DEST=|" mibs/conf/ianarfc.conf;
sed -i "s|^DIR=/usr/share/doc|DIR=\$HOME/php/mibs/originals|" mibs/conf/rfc.conf;
sed -i "s|^DEST=ietf|DEST=|" mibs/conf/rfc.conf;
sed -i "s|^BASEDIR=/var/lib/mibs|BASEDIR=\$HOME/php/mibs|" mibs/conf/snmp-mibs-downloader.conf;
`

	err := RunCommands(command, s.phpDir, nil)
	if err != nil {
		return err
	}

	return nil
}

type IonCube struct {
	data       ExtData
	phpDir     string
	phpVersion string
}

func (i IonCube) Install() error {
	if i.data.Version == "" { // only php v8.1 needs ioncube
		log.Printf("Ioncube is not included in extensions for php v%s", i.phpVersion)
		return nil
	}

	log.Printf("Installing IonCube v%s\n", i.data.Version)

	workingDir := fmt.Sprintf("ioncube-%s", i.data.Version)
	url := fmt.Sprintf("http://downloads3.ioncube.com/loader_downloads/ioncube_loaders_lin_x86-64_%s.tar.gz", i.data.Version)

	err := DownloadDependency(workingDir, url, i.data.MD5)
	if err != nil {
		return err
	}

	err = i.setupTar(workingDir, i.phpDir, i.phpVersion)
	if err != nil {
		return err
	}

	return nil
}

func (i IonCube) getZtsPath(phpDir string) (string, error) {
	matches, err := filepath.Glob(fmt.Sprintf("%s/lib/php/extensions/no-debug-non-zts-*", phpDir))
	if err != nil {
		return "", fmt.Errorf("error getting zts path: %s", err)
	}

	return matches[0], nil
}

func (i IonCube) setupTar(workingDir, phpDir, phpVersion string) error {
	majorVersion, err := GetMajorVersion(phpVersion)
	if err != nil {
		return err
	}

	ztsPath, err := i.getZtsPath(phpDir)
	if err != nil {
		return err
	}

	err = RunCommands(fmt.Sprintf("cp ioncube_loader_lin_%s.so %s/ioncube.so", majorVersion, ztsPath), workingDir, nil)
	if err != nil {
		return err
	}

	return nil
}

type RabbitMQ struct {
	data ExtData
}

func (rmq RabbitMQ) Install() error {
	log.Printf("Installing RabbitMQ v%s\n", rmq.data.Version)

	workingDir := fmt.Sprintf("rabbitmq-%s", rmq.data.Version)
	url := fmt.Sprintf("https://github.com/alanxz/rabbitmq-c/archive/v%s.tar.gz", rmq.data.Version)

	err := DownloadDependency(workingDir, url, rmq.data.MD5)
	if err != nil {
		return err
	}

	command := `
cmake .;
cmake build .;
cmake -DCMAKE_INSTALL_PREFIX=/usr/local .;
cmake --build . --target install;
`

	err = RunCommands(command, workingDir, nil)
	if err != nil {
		return err
	}

	return nil
}

type Lua struct {
	data ExtData
}

func (lua Lua) Install() error {
	log.Printf("Installing Lua v%s\n", lua.data.Version)

	workingDir := fmt.Sprintf("%s-%s", "lua", lua.data.Version)
	url := fmt.Sprintf("http://www.lua.org/ftp/lua-%s.tar.gz", lua.data.Version)

	err := DownloadDependency(workingDir, url, lua.data.MD5)
	if err != nil {
		return err
	}

	command := "make linux MYCFLAGS=-fPIC; make install"

	err = RunCommands(command, workingDir, nil)
	if err != nil {
		return err
	}

	return nil
}

type HiRedis struct {
	data   ExtData
	phpDir string
}

func (hredis HiRedis) setupTar(workingDir string) error {
	command := fmt.Sprintf("cp -a lib/libhiredis.so* %s/lib/", hredis.phpDir)
	err := RunCommands(command, workingDir, nil)
	if err != nil {
		return err
	}

	return nil
}

func (hredis HiRedis) Install() error {
	log.Printf("Installing HiRedis v%s\n", hredis.data.Version)

	workingDir := fmt.Sprintf("hiredis-%s", hredis.data.Version)
	url := fmt.Sprintf("https://github.com/redis/hiredis/archive/v%s.tar.gz", hredis.data.Version)

	err := DownloadDependency(workingDir, url, hredis.data.MD5)
	if err != nil {
		return err
	}

	command := "make; make install"

	envVars := []string{"LIBRARY_PATH=lib", "PREFIX=./"}
	err = RunCommands(command, workingDir, envVars)
	if err != nil {
		return err
	}

	err = hredis.setupTar(workingDir)
	if err != nil {
		return err
	}

	return nil
}

type LibRdKafka struct {
	data ExtData
}

func (rdk LibRdKafka) Install() error {
	log.Printf("Installing LibRdKafka v%s\n", rdk.data.Version)

	workingDir := fmt.Sprintf("%s-%s", "librdkafka", rdk.data.Version)
	url := fmt.Sprintf("https://github.com/edenhill/librdkafka/archive/v%s.tar.gz", rdk.data.Version)

	err := DownloadDependency(workingDir, url, rdk.data.MD5)
	if err != nil {
		return err
	}

	command := "./configure --prefix=/usr; make; make install"

	err = RunCommands(command, workingDir, nil)
	if err != nil {
		return err
	}

	return nil
}

type LibSodium struct {
	data   ExtData
	phpDir string
}

func (sodium LibSodium) setupTar() error {
	workingDir := fmt.Sprintf("libsodium-%s", sodium.data.Version)

	command := fmt.Sprintf("cp -a /usr/local/lib/libsodium.so* %s/lib/", sodium.phpDir)
	err := RunCommands(command, workingDir, nil)
	if err != nil {
		return err
	}

	return nil
}

func (sodium LibSodium) Install() error {
	log.Printf("Installing LibSodium v%s\n", sodium.data.Version)

	workingDir := fmt.Sprintf("libsodium-%s", sodium.data.Version)
	url := fmt.Sprintf("https://download.libsodium.org/libsodium/releases/libsodium-%s.tar.gz", sodium.data.Version)

	err := DownloadDependency(workingDir, url, sodium.data.MD5)
	if err != nil {
		return err
	}

	command := fmt.Sprintf("./configure --with-php-config=%s/bin/php-config --with-sodium=./; make; make install", sodium.phpDir)
	err = RunCommands(command, workingDir, nil)
	if err != nil {
		return err
	}

	err = sodium.setupTar()
	if err != nil {
		return err
	}

	return nil
}
