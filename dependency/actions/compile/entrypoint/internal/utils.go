package internal

import (
	"crypto/md5"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
)

type Installer interface {
	Install() error
}

type ExtData struct {
	MD5     string
	Version string
}

type PeclData struct {
	Name    string
	MD5     string
	Version string
}

type ExtensionsData struct {
	NativeModules   map[string]ExtData `json:"native_modules"`
	PeclExtensions  []PeclData         `json:"pecl_extensions"`
	LocalExtensions []string           `json:"local_extensions"`
}

func GetMajorVersion(version string) (string, error) {
	r, err := regexp.Compile(`\d+\.\d+`)
	if err != nil {
		return "", err
	}

	return r.FindString(version), nil
}

func getExtensionsData(phpVersion string) (ExtensionsData, error) {
	majorVersion, err := GetMajorVersion(phpVersion)
	if err != nil {
		return ExtensionsData{}, err
	}

	extFile := fmt.Sprintf("/tmp/entrypoint/extensions/extensions-%s.json", majorVersion)
	jsonBlob, err := os.ReadFile(extFile)
	if err != nil {
		return ExtensionsData{}, err
	}

	var extensions ExtensionsData
	err = json.Unmarshal(jsonBlob, &extensions)
	if err != nil {
		return ExtensionsData{}, err
	}

	return extensions, nil
}

func InstallExtensions(version, phpDir string) error {
	extensions, err := getExtensionsData(version)
	if err != nil {
		return err
	}

	installers := []Installer{}
	nativeModules := extensions.NativeModules
	installers = append(installers,
		IonCube{nativeModules["ioncube"], phpDir, version},
		RabbitMQ{nativeModules["rabbitmq"]},
		Lua{nativeModules["lua"]},
		HiRedis{nativeModules["hiredis"], phpDir},
		Snmp{phpDir},
		LibRdKafka{nativeModules["librdkafka"]},
		LibSodium{nativeModules["libsodium"], phpDir},
		TidewaysXhprof{nativeModules["tideways_xhprof"], phpDir},
		Phpiredis{nativeModules["phpiredis"], phpDir, fmt.Sprintf("/hiredis-%s", nativeModules["hiredis"].Version)},
		PeclExtensions{extensions.PeclExtensions, phpDir},
		LocalExtensions{extensions.LocalExtensions, version, phpDir, fmt.Sprintf("/libsodium-%s", nativeModules["libsodium"].Version)},
	)

	for _, installer := range installers {
		err = installer.Install()
		if err != nil {
			return err
		}
	}

	return nil
}

func VerifyMD5(path, expectedMD5 string) error {
	actualMD5, err := getMD5(path)
	if err != nil {
		return fmt.Errorf("failed to get actual MD5: %w", err)
	}

	if actualMD5 != expectedMD5 {
		return fmt.Errorf("expected MD5 '%s' but got '%s'", expectedMD5, actualMD5)
	}

	return nil
}

func getMD5(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "nil", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "nil", fmt.Errorf("failed to calculate MD5: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func DownloadDependency(workingDir, url, sha string) error {
	err := os.Mkdir(workingDir, os.ModePerm)
	if err != nil {
		return err
	}

	command := fmt.Sprintf(`
curl --location %s --silent --output upstream.tgz;
tar --extract --file upstream.tgz --directory %s --strip-components=1`, url, workingDir)

	err = RunCommands(command, "", nil)
	if err != nil {
		return err
	}

	err = VerifyMD5("upstream.tgz", sha)
	if err != nil {
		return err
	}

	return nil
}

func DownloadAndExtract(dep, version, url string) error {
	log.Printf("Downloading upstream tarball for %s v%s", dep, version)

	depDir := fmt.Sprintf("%s-%s", dep, version)
	err := os.Mkdir(depDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create directory for upstream dependency %s", fmt.Sprintf("%s-%s", dep, version))
	}

	command := fmt.Sprintf(`
curl --location %s --silent --output upstream.tgz;
tar --extract --file upstream.tgz --directory %s --strip-components=1`, url, depDir)

	err = RunCommands(command, "", nil)
	if err != nil {
		return err
	}

	return nil
}

func RunCommands(command, workingDir string, envVars []string) error {
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Dir = workingDir
	cmd.Env = append(cmd.Environ(), envVars...)

	if debug() {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func debug() bool {
	return os.Getenv("DEBUG_PHP_COMPILE") == "true"
}
