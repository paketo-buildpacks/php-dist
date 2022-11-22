package retrieval

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/joshuatcasey/libdependency/retrieve"
	"github.com/joshuatcasey/libdependency/upstream"
	"github.com/joshuatcasey/libdependency/versionology"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

// GenerateMetadata returns the dependency metadata for a given version of PHP
func GenerateMetadata(versionFetcher versionology.VersionFetcher) ([]versionology.Dependency, error) {
	logger := scribe.NewEmitter(os.Stdout).
		WithLevel(os.Getenv("BP_LOG_LEVEL"))

	phpReleasePretty, ok := versionFetcher.(PhpReleasePretty)
	if !ok {
		return nil, errors.New("expected a PhpReleasePretty")
	}

	version := phpReleasePretty.Version().String()
	sourceURL := fmt.Sprintf("https://www.php.net/distributions/php-%s.tar.gz", version)

	err := verifyChecksumAndSignature(phpReleasePretty, sourceURL, logger)
	if err != nil {
		return nil, err
	}

	var deprecationDate *time.Time
	if phpReleasePretty.ReleaseDate != nil {
		temp := phpReleasePretty.ReleaseDate.AddDate(3, 0, 0)
		deprecationDate = &temp
	}

	configMetadataDependency := cargo.ConfigMetadataDependency{
		CPE:             fmt.Sprintf("cpe:2.3:a:php:php:%s:*:*:*:*:*:*:*", version),
		DeprecationDate: deprecationDate,
		ID:              "php",
		Licenses:        retrieve.LookupLicenses(sourceURL, upstream.DefaultDecompress),
		Name:            "PHP",
		PURL:            retrieve.GeneratePURL("php", version, phpReleasePretty.Sha256, sourceURL),
		Source:          sourceURL,
		SourceChecksum:  fmt.Sprintf("sha256:%s", phpReleasePretty.Sha256),
		Stacks:          []string{"io.buildpacks.stacks.bionic"},
		Version:         version,
	}

	deps := []versionology.Dependency{
		{
			ConfigMetadataDependency: configMetadataDependency,
			SemverVersion:            versionFetcher.Version(),
			Target:                   "bionic",
		},
	}

	if versionFetcher.Version().GreaterThan(semver.MustParse("8.1.0")) {
		configMetadataDependency := cargo.ConfigMetadataDependency{
			CPE:             fmt.Sprintf("cpe:2.3:a:php:php:%s:*:*:*:*:*:*:*", version),
			DeprecationDate: deprecationDate,
			ID:              "php",
			Licenses:        retrieve.LookupLicenses(sourceURL, upstream.DefaultDecompress),
			Name:            "PHP",
			PURL:            retrieve.GeneratePURL("php", version, phpReleasePretty.Sha256, sourceURL),
			Source:          sourceURL,
			SourceChecksum:  fmt.Sprintf("sha256:%s", phpReleasePretty.Sha256),
			Stacks:          []string{"io.buildpacks.stacks.jammy"},
			Version:         version,
		}
		deps = append(deps, versionology.Dependency{
			ConfigMetadataDependency: configMetadataDependency,
			SemverVersion:            versionFetcher.Version(),
			Target:                   "jammy",
		})

	}
	return deps, nil
}

func hasKey[T comparable, U any](m map[T]U, key T) bool {
	if _, ok := m[key]; ok {
		return true
	}

	return false
}

// GetAllVersions calls php.net to check the active releases for newer versions than those found in buildpack.toml
func GetAllVersions(getPhpReleases GetPhpReleases) (versionology.VersionFetcherArray, error) {
	// Another option: https://github.com/php/php-src/tags
	indexUrl := "https://www.php.net/releases/active"

	indexData, err := getPhpReleases.Get(indexUrl)
	if err != nil {
		return nil, err
	}

	var prettyReleases versionology.VersionFetcherArray

	if !hasKey(indexData, "7") || !hasKey(indexData["7"], "7.4") {
		return nil, errors.New("unable to find releases for PHP 7.4")
	}

	if !hasKey(indexData, "8") || !hasKey(indexData["8"], "8.0") || !hasKey(indexData["8"], "8.1") {
		return nil, errors.New("unable to find releases for PHP 8.0 or 8.1")
	}

	if release, err := makeItPretty(indexData["7"]["7.4"]); err != nil {
		return nil, err
	} else {
		prettyReleases = append(prettyReleases, release)
	}

	if release, err := makeItPretty(indexData["8"]["8.0"]); err != nil {
		return nil, err
	} else {
		prettyReleases = append(prettyReleases, release)
	}

	if release, err := makeItPretty(indexData["8"]["8.1"]); err != nil {
		return nil, err
	} else {
		prettyReleases = append(prettyReleases, release)
	}

	return prettyReleases, nil
}

func makeItPretty(raw PhpReleaseAnnouncementRaw) (PhpReleasePretty, error) {
	var err error
	var parsedDate time.Time

	version, err := semver.NewVersion(raw.Version)
	if err != nil {
		return PhpReleasePretty{}, fmt.Errorf("unable to parse semantic version %s: %w", raw.Version, err)
	}

	if parsedDate, err = time.Parse("02 Jan 2006", raw.Date); err != nil {
		return PhpReleasePretty{}, fmt.Errorf("unable to parse date for PHP version %s: %w", version, err)
	}

	var sha256 string
	for _, source := range raw.Source {
		if !strings.HasSuffix(source.Filename, ".tar.gz") || source.Sha256 == "" {
			continue
		}
		sha256 = source.Sha256
	}

	if sha256 == "" {
		return PhpReleasePretty{}, fmt.Errorf("unable to find SHA256 for PHP version %s", version)
	}

	return PhpReleasePretty{
		SemverVersion: version,
		Sha256:        sha256,
		ReleaseDate:   &parsedDate,
	}, nil
}

type PhpReleasePretty struct {
	SemverVersion *semver.Version
	Sha256        string
	ReleaseDate   *time.Time
}

func (pretty PhpReleasePretty) Version() *semver.Version {
	return pretty.SemverVersion
}
