package phpdist_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit"
	phpdist "github.com/paketo-buildpacks/php-dist"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testEnvironment(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path        string
		layer       packit.Layer
		buffer      *bytes.Buffer
		environment phpdist.Environment
	)

	it.Before(func() {
		var err error
		path, err = ioutil.TempDir("", "layer-dir")
		Expect(err).NotTo(HaveOccurred())

		layer = packit.Layer{Path: path}

		err = layer.Reset()
		Expect(err).NotTo(HaveOccurred())

		err = os.MkdirAll(filepath.Join(layer.Path, "lib/php/extensions/no-debug-non-zts-20200717"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		buffer = bytes.NewBuffer(nil)
		environment = phpdist.NewEnvironment(phpdist.NewLogEmitter(buffer))
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("Configure", func() {
		it("configures the environment variables", func() {
			err := environment.Configure(layer)
			Expect(err).NotTo(HaveOccurred())

			Expect(layer.SharedEnv).To(Equal(packit.Environment{
				"MIBDIRS.override":           filepath.Join(layer.Path, "mibs"),
				"PATH.delim":                 ":",
				"PATH.prepend":               filepath.Join(layer.Path, "sbin"),
				"PHP_API.override":           "20200717",
				"PHP_EXTENSION_DIR.override": filepath.Join(layer.Path, "lib/php/extensions/no-debug-non-zts-20200717"),
				"PHP_HOME.override":          layer.Path,
			}))
		})
	})

	context("extensions dir does not exist", func() {
		var err error

		it.Before(func() {
			err = os.RemoveAll(filepath.Join(layer.Path, "lib/php/extensions/no-debug-non-zts-20200717"))
			Expect(err).NotTo(HaveOccurred())
		})

		it("throws a descriptive error", func() {
			err := environment.Configure(layer)
			Expect(err).To(MatchError(ContainSubstring("php extensions dir not found")))
		})
	})
}
