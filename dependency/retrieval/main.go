package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/joshuatcasey/libdependency/retrieve"
	"github.com/joshuatcasey/libdependency/upstream"
	"github.com/joshuatcasey/libdependency/versionology"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

type PhpMetadata struct {
	SemverVersion *semver.Version
}

func (phpMetadata PhpMetadata) Version() *semver.Version {
	return phpMetadata.SemverVersion
}

type PhpSource struct {
	Filename string `json:"filename"`
	SHA256   string `json:"sha256"`
	MD5      string `json:"md5"`
}

type PhpRelease struct {
	Version string
	Date    *time.Time
	Source  []PhpSource
}

type PhpRawRelease struct {
	Date   string      `json:"date"`
	Source []PhpSource `json:"source"`
	Museum bool        `json:"museum"`
}

func main() {
	retrieve.NewMetadata("php", getAllVersions, generateMetadata)
}

func getAllVersions() (versionology.VersionFetcherArray, error) {
	phpReleases, err := getPhpReleases()
	if err != nil {
		return nil, err
	}

	sort.Slice(phpReleases, func(i, j int) bool {
		if phpReleases[i].Date.Equal(*phpReleases[j].Date) {
			return phpReleases[i].Version > phpReleases[j].Version
		}
		return phpReleases[i].Date.After(*phpReleases[j].Date)
	})

	var versions []versionology.VersionFetcher
	for _, release := range phpReleases {
		versions = append(versions, PhpMetadata{
			semver.MustParse(release.Version),
		})
	}

	return versions, nil
}

func generateMetadata(versionFetcher versionology.VersionFetcher) ([]versionology.Dependency, error) {
	version := versionFetcher.Version().String()

	release, err := getRelease(version)
	if err != nil {
		return nil, fmt.Errorf("could not get release: %w", err)
	}

	dependencyURL := dependencyURL(release, version)
	dependencySHA, err := getDependencySHA(release, version)
	if err != nil {
		return nil, err
	}

	// // releaseDate is the patch releaseDate
	// releaseDate, err := getReleaseDate(release)
	// if err != nil {
	// 	return nil, fmt.Errorf("could not get release date: %w", err)
	// }

	// deprecationDate is the minor version line deprecation date
	deprecationDate, err := getDeprecationDate(version)
	if err != nil {
		return nil, fmt.Errorf("could not get version line deprecation date: %w", err)
	}

	dep := cargo.ConfigMetadataDependency{
		Version:         version,
		ID:              "php",
		Name:            "PHP",
		Source:          dependencyURL,
		SourceChecksum:  fmt.Sprintf("sha256:%s", dependencySHA),
		CPE:             fmt.Sprintf("cpe:2.3:a:php:php:%s:*:*:*:*:*:*:*", version),
		PURL:            retrieve.GeneratePURL("php", version, dependencySHA, dependencyURL),
		Licenses:        retrieve.LookupLicenses(dependencyURL, upstream.DefaultDecompress),
		DeprecationDate: deprecationDate,
		Stacks:          []string{"io.buildpacks.stacks.bionic"},
	}

	bionicDependency, err := versionology.NewDependency(dep, "bionic")
	if err != nil {
		return nil, fmt.Errorf("could get create bionic dependency: %w", err)
	}

	dependencies := []versionology.Dependency{bionicDependency}

	// If target==jammy and version >= 8.1, include it
	// Versions less than 8.1 are not supported on Jammy.
	semVersion, err := semver.NewVersion(version)
	if err != nil {
		return nil, err
	}
	constraint, err := semver.NewConstraint(">= 8.1")
	if err != nil {
		//untested
		return nil, err
	}

	if constraint.Check(semVersion) {
		dep.Stacks = []string{"io.buildpacks.stacks.jammy"}

		jammyDependency, err := versionology.NewDependency(dep, "jammy")
		if err != nil {
			return nil, fmt.Errorf("could get create jammy dependency: %w", err)
		}

		dependencies = append(dependencies, jammyDependency)
	}

	return dependencies, nil
}

func getPhpReleases() ([]PhpRelease, error) {
	webClient := NewWebClient()
	body, err := webClient.Get("https://raw.githubusercontent.com/brayanhenao/php-releases-information/main/releases.json")
	if err != nil {
		return nil, fmt.Errorf("could not hit php.net: %w", err)
	}

	var phpLines map[string]interface{}
	err = json.Unmarshal(body, &phpLines)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal version lines response: %w\n%s", err, body)
	}

	var versionLines []string
	for line := range phpLines {
		if line == "3" {
			continue
		}

		versionLines = append(versionLines, line)
	}
	sort.Strings(versionLines)

	var allPhpReleases []PhpRelease

	for _, line := range versionLines {
		body, err = webClient.Get(fmt.Sprintf("https://raw.githubusercontent.com/brayanhenao/php-releases-information/main/php-%s.json", line))
		if err != nil {
			return nil, fmt.Errorf("could not hit php.net: %w", err)
		}

		var phpRawReleases map[string]PhpRawRelease
		err = json.Unmarshal(body, &phpRawReleases)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal version lines response: %w\n%s", err, body)
		}

		for version, release := range phpRawReleases {
			releaseDate, err := parseReleaseDate(release.Date)
			if err != nil {
				return nil, fmt.Errorf("could not parse release date: %w", err)
			}

			allPhpReleases = append(allPhpReleases, PhpRelease{
				Version: version,
				Date:    releaseDate,
				Source:  release.Source,
			})
		}
	}

	return allPhpReleases, nil
}

func parseReleaseDate(date string) (*time.Time, error) {
	if parsedDate, err := time.Parse("02 Jan 2006", date); err == nil {
		return &parsedDate, nil
	}

	if parsedDate, err := time.Parse("2 Jan 2006", date); err == nil {
		return &parsedDate, nil
	}

	if parsedDate, err := time.Parse("02 January 2006", date); err == nil {
		return &parsedDate, nil
	}

	if parsedDate, err := time.Parse("2 January 2006", date); err == nil {
		return &parsedDate, nil
	}

	return nil, fmt.Errorf("release date '%s' did not match any expected patterns", date)
}

func getRelease(version string) (PhpRawRelease, error) {
	webClient := NewWebClient()
	semverSplit := strings.Split(version, ".")

	searchMajorVersion := semverSplit[0]
	patchVersion := semverSplit[2]

	// Mirroring what PHP does, it converts the wildcard patch version to the oldest patch version for that line.
	// Eg:  7.4.x ---- 7.4.0
	// Note: Assuming that the oldest patch version is always 0.
	if patchVersion == "*" {
		version = strings.ReplaceAll(version, "*", "0")
	}

	body, err := webClient.Get(fmt.Sprintf("https://raw.githubusercontent.com/brayanhenao/php-releases-information/main/php-%s.json", searchMajorVersion))
	if err != nil {
		fmt.Println(string(body))
		return PhpRawRelease{}, fmt.Errorf("could not hit php.net: %w", err)
	}

	var phpRawReleases map[string]PhpRawRelease
	err = json.Unmarshal(body, &phpRawReleases)
	if err != nil {
		return PhpRawRelease{}, fmt.Errorf("could not unmarshal version lines response: %w\n%s", err, body)
	}

	for rawPhpVersion, release := range phpRawReleases {
		if rawPhpVersion == version {
			return release, nil
		}
	}

	return PhpRawRelease{}, nil
}

func dependencyURL(release PhpRawRelease, version string) string {
	if release.Museum {
		majorVersion := version[0:1]
		return fmt.Sprintf("https://museum.php.net/php%s/php-%s.tar.gz", majorVersion, version)
	}

	return fmt.Sprintf("https://github.com/php/web-php-distributions/raw/master/php-%s.tar.gz", version)
}

func getDependencySHA(release PhpRawRelease, version string) (string, error) {
	for _, file := range release.Source {
		if filepath.Ext(file.Filename) == ".gz" {
			if file.SHA256 != "" {
				return file.SHA256, nil
			} else if file.MD5 != "" || dependencyVersionIsMissingChecksum(version) {
				sha, err := getSHA256FromReleaseFile(release, file, version)
				if err != nil {
					return "", fmt.Errorf("could not get SHA256 from release file: %w", err)
				}

				return sha, nil
			} else {
				return "", fmt.Errorf("could not find SHA256 or MD5 for %s", version)
			}
		}
	}

	return "", fmt.Errorf("could not find .tar.gz file for %s", version)
}

func dependencyVersionIsMissingChecksum(version string) bool {
	versionsWithMissingChecksum := map[string]bool{
		"5.1.6":  true,
		"5.1.5":  true,
		"5.1.4":  true,
		"5.1.3":  true,
		"5.1.2":  true,
		"5.1.1":  true,
		"5.1.0":  true,
		"5.0.5":  true,
		"5.0.4":  true,
		"5.0.3":  true,
		"5.0.2":  true,
		"5.0.1":  true,
		"5.0.0":  true,
		"4.4.5":  true,
		"4.4.4":  true,
		"4.4.3":  true,
		"4.4.2":  true,
		"4.4.1":  true,
		"4.4.0":  true,
		"4.3.11": true,
		"4.3.10": true,
		"4.3.9":  true,
		"4.3.8":  true,
		"4.3.7":  true,
		"4.3.6":  true,
		"4.3.5":  true,
		"4.3.4":  true,
		"4.3.3":  true,
		"4.3.2":  true,
		"4.3.1":  true,
		"4.3.0":  true,
		"4.2.3":  true,
		"4.2.2":  true,
		"4.2.1":  true,
		"4.2.0":  true,
		"4.1.2":  true,
		"4.1.1":  true,
		"4.1.0":  true,
		"4.0.6":  true,
		"4.0.5":  true,
		"4.0.4":  true,
		"4.0.3":  true,
		"4.0.2":  true,
		"4.0.1":  true,
		"4.0.0":  true,
	}

	_, shouldBeIgnored := versionsWithMissingChecksum[version]
	return shouldBeIgnored
}

func getSHA256FromReleaseFile(release PhpRawRelease, file PhpSource, version string) (string, error) {
	webClient := NewWebClient()
	tempDir, err := os.MkdirTemp("", "php")
	if err != nil {
		return "", fmt.Errorf("could not create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	dependencyOutputPath := filepath.Join(tempDir, file.Filename)
	err = webClient.Download(dependencyURL(release, version), dependencyOutputPath)
	if err != nil {
		return "", fmt.Errorf("could not download dependency: %w", err)
	}

	if !dependencyVersionHasIncorrectChecksum(version) && file.MD5 != "" {
		err = verifyMD5(dependencyOutputPath, file.MD5)
		if err != nil {
			return "", fmt.Errorf("dependency signature verification failed: %w", err)
		}
	}

	sha256, err := getSHA256(dependencyOutputPath)
	if err != nil {
		return "", fmt.Errorf("could not get SHA256: %w", err)
	}

	return sha256, nil
}

func dependencyVersionHasIncorrectChecksum(version string) bool {
	versionsWithWrongChecksum := map[string]bool{
		"5.3.25": true,
		"5.3.11": true,
		"5.2.14": true,
	}

	_, shouldBeIgnored := versionsWithWrongChecksum[version]
	return shouldBeIgnored
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

func verifyMD5(path, expectedMD5 string) error {
	actualMD5, err := getMD5(path)
	if err != nil {
		return fmt.Errorf("failed to get actual MD5: %w", err)
	}

	if actualMD5 != expectedMD5 {
		return fmt.Errorf("expected MD5 '%s' but got '%s'", expectedMD5, actualMD5)
	}

	return nil
}

func getSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "nil", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "nil", fmt.Errorf("failed to calculate SHA256: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func getReleaseDate(release PhpRawRelease) (*time.Time, error) {
	if parsedDate, err := time.Parse("02 Jan 2006", release.Date); err == nil {
		return &parsedDate, nil
	}

	if parsedDate, err := time.Parse("2 Jan 2006", release.Date); err == nil {
		return &parsedDate, nil
	}

	if parsedDate, err := time.Parse("02 January 2006", release.Date); err == nil {
		return &parsedDate, nil
	}

	if parsedDate, err := time.Parse("2 January 2006", release.Date); err == nil {
		return &parsedDate, nil
	}

	return nil, fmt.Errorf("release date '%s' did not match any expected patterns", release.Date)
}

func getDeprecationDate(version string) (*time.Time, error) {
	semVer, err := semver.NewVersion(version)
	if err != nil {
		return nil, fmt.Errorf("could not calculate minor version line for %s: %w", version, err)
	}

	versionLine := fmt.Sprintf("%d.%d.*", semVer.Major(), semVer.Minor())
	release, err := getRelease(versionLine)
	if err != nil {
		return nil, fmt.Errorf("could not get version-line release: %w", err)
	}
	versionLineReleaseDate, err := getReleaseDate(release)
	if err != nil {
		return nil, fmt.Errorf("could not get version-line release: %w", err)
	}
	return calculateDeprecationDate(*versionLineReleaseDate), nil
}

func calculateDeprecationDate(releaseDate time.Time) *time.Time {
	deprecationDate := time.Date(releaseDate.Year()+3, releaseDate.Month(), releaseDate.Day(),
		0, 0, 0, 0, time.UTC)

	return &deprecationDate
}
