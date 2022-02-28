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

		containerIDs := map[string]interface{}{}

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			for id := range containerIDs {
				Expect(docker.Container.Remove.Execute(id)).To(Succeed())
			}
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
				WithPullPolicy("never").
				WithBuildpacks(
					phpDistBuildpack,
					buildPlanBuildpack,
				).
				WithEnv(map[string]string{
					"BP_LOG_LEVEL": "DEBUG",
				}).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String())

			Expect(logs).To(ContainLines(
				fmt.Sprintf("%s %s", buildpackInfo.Buildpack.Name, version),
				"  Resolving PHP version",
				"    Candidate version sources (in priority order):",
				"      buildpack.yml -> \"8.0.*\"",
				"      <unknown>     -> \"\"",
				"",
				MatchRegexp(`    Selected PHP version \(using buildpack\.yml\): 8\.0\.\d+`),
				"",
				MatchRegexp(fmt.Sprintf(`    WARNING: Setting the PHP version through buildpack.yml will be deprecated in %s v\d+\.\d+\.\d+\.`, buildpackInfo.Buildpack.Name)),
				MatchRegexp(`    In versions >= v\d+\.\d+\.\d+, use the \$BP_PHP_VERSION environment variable to specify a version\.`),
				"",
				"  Getting the layer associated with PHP:",
				fmt.Sprintf("    /layers/%s/php", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
				"",
				"  Generating the SBOM",
				"",
				"  Executing build process",
				MatchRegexp(`    Installing PHP 8\.0\.\d+`),
				fmt.Sprintf("    Installation path: /layers/%s/php", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
				MatchRegexp(`    Dependency URI: https:\/\/deps\.paketo.io\/php\/.*\.tgz`),
				MatchRegexp(`      Completed in \d+\.\d+`),
				"",
				"    Finding PHP extensions directory",
				"",
				"    Generating default PHP configuration",
				fmt.Sprintf("      Generated /layers/%[1]s/php/etc/php.ini and /layers/%[1]s/php/etc/buildpack.ini", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
				"",
			))
			Expect(logs).To(ContainLines("  Configuring build environment"))
			Expect(logs).To(ContainLines("  Configuring launch environment"))
			Expect(logs).To(ContainLines("  PHP layer will be available at runtime"))

			container, err = docker.Container.Run.WithCommand("php -i && sleep infinity").Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())
			containerIDs[container.ID] = struct{}{}

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(container.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(
				And(
					MatchRegexp(`PHP Version => 8\.0\.\d+`),
					ContainSubstring(
						fmt.Sprintf("PHP_HOME => /layers/%s/php", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
					),
					ContainSubstring(
						fmt.Sprintf("MIBDIRS => /layers/%s/php/mibs", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
					),
					MatchRegexp(
						fmt.Sprintf(`PHP_EXTENSION_DIR => /layers/%s/php/lib/php/extensions/no-debug-non-zts-\d+`, strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
					),
					ContainSubstring(
						fmt.Sprintf("PHPRC => /layers/%s/php/etc", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
					),
					ContainSubstring(
						fmt.Sprintf("PHP_INI_SCAN_DIR => /layers/%s/php/etc", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
					),
				),
			)

			container, err = docker.Container.Run.WithCommand("php --ini && sleep infinity").Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())
			containerIDs[container.ID] = struct{}{}

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(container.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(
				And(
					MatchRegexp(
						fmt.Sprintf(`Loaded Configuration File:\s+%s`, filepath.Join("/layers", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "php", "etc", "php.ini")),
					),
					MatchRegexp(
						fmt.Sprintf(`Scan for additional .ini files in:\s+%s`, filepath.Join("/layers", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "php", "etc")),
					),
					MatchRegexp(
						fmt.Sprintf(`Additional .ini files parsed:\s+%s`, filepath.Join("/layers", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "php", "etc", "buildpack.ini")),
					),
				),
			)

			container, err = docker.Container.Run.WithCommand(fmt.Sprintf("cat %s", filepath.Join("/layers", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "php", "etc", "buildpack.ini"))).Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())
			containerIDs[container.ID] = struct{}{}

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(container.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(
				And(
					ContainSubstring(
						fmt.Sprintf(`include_path = "%s:/workspace/lib"`, filepath.Join("/layers", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "php", "lib", "php")),
					),
				),
			)
		})
	})
}
