package main

import (
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/detect"

	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDetect(t *testing.T) {
	RegisterTestingT(t)
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, when spec.G, it spec.S) {
	var factory *test.DetectFactory

	it.Before(func() {

		factory = test.NewDetectFactory(t)
	})

	when("there is an php app", func() {
		it("should pass with the default version of php", func() {
			test.WriteFile(t, filepath.Join(factory.Detect.Application.Root, "htdocs", "index.php"), "")
			code, err := runDetect(factory.Detect)
			Expect(err).NotTo(HaveOccurred())

			Expect(code).To(Equal(detect.PassStatusCode))
		})

		it("should pass with the default version of php", func() {
			test.WriteFile(t, filepath.Join(factory.Detect.Application.Root, "htdocs", "my_cool_app.php"), "")
			code, err := runDetect(factory.Detect)
			Expect(err).NotTo(HaveOccurred())

			Expect(code).To(Equal(detect.PassStatusCode))
		})
	})
}
