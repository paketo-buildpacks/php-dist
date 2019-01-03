package php

import (
	"fmt"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/sclevine/spec/report"

	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func TestUnitPHP(t *testing.T) {
	RegisterTestingT(t)
	spec.Run(t, "PHP", testPHP, spec.Report(report.Terminal{}))
}

func testPHP(t *testing.T, when spec.G, it spec.S) {
	when("NewContributor", func() {
		var stubPHPFixture = filepath.Join("fixtures", "stub-php.tar.gz")

		it("returns true if a build plan exists", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(Dependency, buildplan.Dependency{})
			f.AddDependency(Dependency, stubPHPFixture)

			_, willContribute, err := NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeTrue())
		})

		it("returns false if a build plan does not exist", func() {
			f := test.NewBuildFactory(t)

			_, willContribute, err := NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeFalse())
		})

		it("should contribute php to launch when launch is true", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{"launch": true},
			})
			f.AddDependency(Dependency, stubPHPFixture)

			phpContributor, _, err := NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(phpContributor.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer(Dependency)
			root := f.Build.Application.Root

			Expect(layer).To(test.HaveLayerMetadata(false, true, true))

			Expect(layer).To(test.HaveOverrideSharedEnvironment("PHPRC", "%s/etc", layer.Root))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("MIBDIRS", "%s/mibs", layer.Root))

			Expect(layer).To(test.HaveOverrideSharedEnvironment("PHP_INI_SCAN_DIR", "%s/etc/php.ini.d", root))
			Expect(filepath.Join(layer.Root, "stub.txt")).To(BeARegularFile())
			Expect(f.Build.Layers).To(test.HaveLaunchMetadata(
				layers.Metadata{Processes: []layers.Process{{"web", fmt.Sprintf("php -S 0.0.0.0:8080 -t %s/htdocs", root)}}},
			))
		})

		it("should contribute php to build when build is true", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{"build": true},
			})
			f.AddDependency(Dependency, stubPHPFixture)

			phpContributor, _, err := NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(phpContributor.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer(Dependency)

			Expect(layer).To(test.HaveLayerMetadata(true, true, false))

			Expect(layer).To(test.HaveOverrideSharedEnvironment("PHPRC", "%s/etc", layer.Root))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("MIBDIRS", "%s/mibs", layer.Root))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("PHP_INI_SCAN_DIR", "%s/etc/php.ini.d", f.Build.Application.Root))

			Expect(filepath.Join(layer.Root, "stub.txt")).To(BeARegularFile())
		})

	})
}
