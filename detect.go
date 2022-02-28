package phpdist

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
)

//go:generate faux --interface VersionParser --output fakes/version_parser.go

// VersionParser defines the interface for determining the version of php.
type VersionParser interface {
	ParseVersion(path string) (version string, err error)
}

// BuildPlanMetadata is the buildpack specific data included in build plan
// requirements.
type BuildPlanMetadata struct {
	Version       string `toml:"version"`
	VersionSource string `toml:"version-source"`
}

// Detect will return a packit.DetectFunc that will be invoked during the
// detect phase of the buildpack lifecycle.
//
// Detect always passes, and will contribute a Build Plan that provides php.
func Detect(buildpackYMLParser VersionParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		var requirements []packit.BuildPlanRequirement

		version := os.Getenv("BP_PHP_VERSION")
		if version != "" {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "php",
				Metadata: BuildPlanMetadata{
					Version:       version,
					VersionSource: "BP_PHP_VERSION",
				},
			})
		}

		var err error
		version, err = buildpackYMLParser.ParseVersion(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil {
			return packit.DetectResult{}, err
		}

		if version != "" {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "php",
				Metadata: BuildPlanMetadata{
					Version:       version,
					VersionSource: "buildpack.yml",
				},
			})
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "php"},
				},
				Requires: requirements,
			},
		}, nil
	}
}
