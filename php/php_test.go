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

			nodeContributor, _, err := NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(nodeContributor.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer(Dependency)

			Expect(layer).To(test.HaveLayerMetadata(false, false, true))
			Expect(layer).To(test.HaveOverrideLaunchEnvironment("APP_ROOT", f.Build.Application.Root))
			Expect(layer).To(test.HaveOverrideLaunchEnvironment("SERVER_ROOT", layer.Root))
			Expect(filepath.Join(layer.Root, "stub.txt")).To(BeARegularFile())
			Expect(f.Build.Layers).To(test.HaveLaunchMetadata(
				layers.Metadata{Processes: []layers.Process{{"web", fmt.Sprintf("httpd -f %s -k start -DFOREGROUND", filepath.Join(f.Build.Application.Root, "httpd.conf"))}}},
			))
		})
	})
}
