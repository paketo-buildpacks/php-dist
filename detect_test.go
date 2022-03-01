package phpdist_test

import (
	"errors"
	"os"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	phpdist "github.com/paketo-buildpacks/php-dist"
	"github.com/paketo-buildpacks/php-dist/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir         string
		buildpackYMLParser *fakes.VersionParser
		detect             packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		buildpackYMLParser = &fakes.VersionParser{}
		detect = phpdist.Detect(buildpackYMLParser)
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("returns a plan that provides php", func() {
		it("detects", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "php"},
				},
			}))
		})
	})

	context("when $BP_PHP_VERSION is set and there is a buildpack.yml-defined version", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_PHP_VERSION", "1.2.3")).To(Succeed())
			buildpackYMLParser.ParseVersionCall.Returns.Version = "0.2.4"
		})
		it.After(func() {
			Expect(os.Unsetenv("BP_PHP_VERSION")).To(Succeed())
		})

		it("returns a plan that provides and requires both versions of PHP", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "php"},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "php",
						Metadata: phpdist.BuildPlanMetadata{
							Version:       "1.2.3",
							VersionSource: "BP_PHP_VERSION",
						},
					},
					{
						Name: "php",
						Metadata: phpdist.BuildPlanMetadata{
							Version:       "0.2.4",
							VersionSource: "buildpack.yml",
						},
					},
				},
			}))

			Expect(buildpackYMLParser.ParseVersionCall.Receives.Path).To(Equal("/working-dir/buildpack.yml"))
		})
	})

	context("when $BP_PHP_VERSION is set", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_PHP_VERSION", "1.2.3")).To(Succeed())
		})
		it.After(func() {
			Expect(os.Unsetenv("BP_PHP_VERSION")).To(Succeed())
		})

		it("returns a plan that provides and requires the specified versions of PHP", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "php"},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "php",
						Metadata: phpdist.BuildPlanMetadata{
							Version:       "1.2.3",
							VersionSource: "BP_PHP_VERSION",
						},
					},
				},
			}))
		})
	})

	context("when the source code contains a buildpack.yml file", func() {
		it.Before(func() {
			buildpackYMLParser.ParseVersionCall.Returns.Version = "0.2.4"
		})

		it("returns a plan that provides and requires that version of php", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "php"},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "php",
						Metadata: phpdist.BuildPlanMetadata{
							Version:       "0.2.4",
							VersionSource: "buildpack.yml",
						},
					},
				},
			}))

			Expect(buildpackYMLParser.ParseVersionCall.Receives.Path).To(Equal("/working-dir/buildpack.yml"))
		})
	})

	context("failure cases", func() {
		context("when the buildpack.yml parser fails", func() {
			it.Before(func() {
				buildpackYMLParser.ParseVersionCall.Returns.Err = errors.New("failed to parse buildpack.yml")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: "/working-dir",
				})
				Expect(err).To(MatchError("failed to parse buildpack.yml"))
			})
		})
	})
}
