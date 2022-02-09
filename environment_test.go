package phpdist_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	phpdist "github.com/paketo-buildpacks/php-dist"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testEnvironment(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path        string
		layer       packit.Layer
		environment phpdist.Environment
	)

	it.Before(func() {
		var err error
		path, err = os.MkdirTemp("", "layer-dir")
		Expect(err).NotTo(HaveOccurred())

		layer = packit.Layer{Path: path}

		layer, err = layer.Reset()
		Expect(err).NotTo(HaveOccurred())

		err = os.MkdirAll(filepath.Join(layer.Path, "lib/php/extensions/no-debug-non-zts-20200717"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		environment = phpdist.NewEnvironment()
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("Configure", func() {
		it("configures the environment variables", func() {
			err := environment.Configure(layer, "directory/extensions/no-debug-non-zts-20200717", "some/directory/php.ini", []string{"app-root-dir/php.ini.d", "other/directory"})
			Expect(err).NotTo(HaveOccurred())

			Expect(layer.SharedEnv).To(Equal(packit.Environment{
				"MIBDIRS.default":           filepath.Join(layer.Path, "mibs"),
				"PATH.delim":                ":",
				"PATH.prepend":              filepath.Join(layer.Path, "sbin"),
				"PHP_API.default":           "20200717",
				"PHP_EXTENSION_DIR.default": "directory/extensions/no-debug-non-zts-20200717",
				"PHP_HOME.default":          layer.Path,
				"PHPRC.default":             "some/directory",
				"PHP_INI_SCAN_DIR.append":   "app-root-dir/php.ini.d:other/directory",
				"PHP_INI_SCAN_DIR.delim":    ":",
			}))
		})

		context("the PHP_INI_SCAN_DIR is set by the user", func() {
			it.Before(func() {
				Expect(os.Setenv("PHP_INI_SCAN_DIR", "user-scan-dir")).To(Succeed())
			})
			it.After(func() {
				Expect(os.Unsetenv("PHP_INI_SCAN_DIR")).To(Succeed())
			})

			it("configures the environment variables", func() {
				err := environment.Configure(layer, "directory/extensions/no-debug-non-zts-20200717", "some/directory/php.ini", []string{})
				Expect(err).NotTo(HaveOccurred())

				Expect(layer.SharedEnv).To(Equal(packit.Environment{
					"MIBDIRS.default":           filepath.Join(layer.Path, "mibs"),
					"PATH.delim":                ":",
					"PATH.prepend":              filepath.Join(layer.Path, "sbin"),
					"PHP_API.default":           "20200717",
					"PHP_EXTENSION_DIR.default": "directory/extensions/no-debug-non-zts-20200717",
					"PHP_HOME.default":          layer.Path,
					"PHPRC.default":             "some/directory",
					"PHP_INI_SCAN_DIR.default":  "user-scan-dir",
				}))
			})
		})
	})
}
