package main

import (
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/test"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitBuild(t *testing.T) {
	RegisterTestingT(t)
	spec.Run(t, "Build", testBuild, spec.Report(report.Terminal{}))
}

func testBuild(t *testing.T, _ spec.G, it spec.S) {
	it("Succeeds", func() {
		f := test.NewBuildFactory(t)

		code, err := runBuild(f.Build)
		Expect(err).NotTo(HaveOccurred())

		Expect(code).To(Equal(0))
	})
}
