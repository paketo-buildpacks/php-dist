package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testSimpleApp(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually
		pack       occam.Pack
		docker     occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when building a simple app", func() {
		var (
			image     occam.Image
			container occam.Container
			name      string
			source    string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("creates a working OCI image with specified version of php on PATH", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "simple_app"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithNoPull().
				WithBuildpacks(
					phpDistBuildpack,
					buildPlanBuildpack,
				).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String())

			Expect(logs).To(ContainLines(
				fmt.Sprintf("%s %s", buildpackInfo.Buildpack.Name, version),
				"  Resolving PHP version",
				"    Candidate version sources (in priority order):",
				"      buildpack.yml -> \"7.2.*\"",
				"      <unknown>     -> \"\"",
				"",
				MatchRegexp(`    Selected PHP version \(using buildpack\.yml\): 7\.2\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing PHP 7\.2\.\d+`),
				MatchRegexp(`      Completed in \d+\.\d+`),
				"",
				"  Configuring environment",
				fmt.Sprintf(`    MIBDIRS           -> "/layers/%s/php/mibs"`, strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
				fmt.Sprintf(`    PATH              -> "/layers/%s/php/sbin:$PATH"`, strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
				MatchRegexp(`    PHP_API           -> "\d+"`),
				MatchRegexp(fmt.Sprintf(`    PHP_EXTENSION_DIR -> "/layers/%s/php/lib/php/extensions/no-debug-non-zts-\d+"`, strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))),
				fmt.Sprintf(`    PHP_HOME          -> "/layers/%s/php"`, strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
			))

			container, err = docker.Container.Run.WithCommand("php -i && sleep infinity").Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(container.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(
				And(
					MatchRegexp(`PHP Version => 7\.2\.\d+`),
					ContainSubstring(
						fmt.Sprintf("PHP_HOME => /layers/%s/php", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
					),
					ContainSubstring(
						fmt.Sprintf("MIBDIRS => /layers/%s/php/mibs", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
					),
					MatchRegexp(
						fmt.Sprintf(`PHP_EXTENSION_DIR => /layers/%s/php/lib/php/extensions/no-debug-non-zts-\d+`, strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
					),
				),
			)
		})
	})
}
