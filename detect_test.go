package phpdist_test

import (
	"os"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	phpdist "github.com/paketo-buildpacks/php-dist"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string
		detect     packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		detect = phpdist.Detect()
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

	context("when $BP_PHP_VERSION is set", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_PHP_VERSION", "1.2.3")).To(Succeed())
		})
		it.After(func() {
			Expect(os.Unsetenv("BP_PHP_VERSION")).To(Succeed())
		})

		it("returns a plan that provides and requires specified version of PHP", func() {
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
}
