package phpdist_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	phpdist "github.com/paketo-buildpacks/php-dist"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPHPFileManager(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layerDir string
		cnbDir   string
		files    phpdist.PHPFileManager
	)

	it.Before(func() {
		var err error
		layerDir, err = os.MkdirTemp("", "layer")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chmod(layerDir, os.ModePerm)).To(Succeed())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chmod(cnbDir, os.ModePerm)).To(Succeed())

		Expect(os.MkdirAll(filepath.Join(cnbDir, "config"), os.ModePerm)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(cnbDir, "config", "default.ini"), nil, os.ModePerm)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(cnbDir, "config", "buildpack.ini"), []byte(`
include_path = "{{ .IncludePath }}"
extension_dir = "{{ .ExtensionDir }}"
{{ range $extension := .Extensions }}
extension = {{ $extension }}.so
{{ end }}
{{ range $zend_extension := .ZendExtensions }}
zend_extension = {{ $zend_extension }}.so
{{ end }}
`), os.ModePerm)).To(Succeed())

		files = phpdist.NewPHPFileManager()
	})

	it.After(func() {
		Expect(os.RemoveAll(layerDir)).To(Succeed())
	})

	context("WriteConfig", func() {
		it("writes a php.ini file into layerDir based on the template", func() {
			defaultConfig, buildpackConfig, err := files.WriteConfig(layerDir, cnbDir, phpdist.PhpIniConfig{
				IncludePath:  "some/include/path",
				ExtensionDir: "some/extension/dir",
				Extensions: []string{
					"some-ext",
					"other-ext",
				},
				ZendExtensions: []string{
					"some-zend",
					"other-zend",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(defaultConfig).To(Equal(filepath.Join(layerDir, "etc", "php.ini")))
			Expect(buildpackConfig).To(Equal(filepath.Join(layerDir, "etc", "buildpack.ini")))

			Expect(filepath.Join(layerDir, "etc")).To(BeADirectory())
			Expect(filepath.Join(layerDir, "etc", "php.ini")).To(BeARegularFile())
			Expect(filepath.Join(layerDir, "etc", "buildpack.ini")).To(BeARegularFile())

			contents, err := os.ReadFile(filepath.Join(layerDir, "etc", "buildpack.ini"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(ContainSubstring(`include_path = "some/include/path"`))
			Expect(string(contents)).To(ContainSubstring(`extension_dir = "some/extension/dir"`))
			Expect(string(contents)).To(ContainSubstring(`extension = some-ext.so`))
			Expect(string(contents)).To(ContainSubstring(`extension = other-ext.so`))
			Expect(string(contents)).To(ContainSubstring(`zend_extension = some-zend.so`))
			Expect(string(contents)).To(ContainSubstring(`zend_extension = other-zend.so`))
		})

		context("failure cases", func() {
			context("when etc directory cannot be created", func() {
				it.Before(func() {
					Expect(os.Chmod(layerDir, 0000)).To(Succeed())
				})
				it.After(func() {
					Expect(os.Chmod(layerDir, os.ModePerm)).To(Succeed())
				})
				it("returns an error", func() {
					_, _, err := files.WriteConfig(layerDir, cnbDir, phpdist.PhpIniConfig{})
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
			context("when default.ini cannot be copied", func() {
				it.Before(func() {
					Expect(os.MkdirAll(filepath.Join(layerDir, "etc"), os.ModePerm)).To(Succeed())
					Expect(os.Chmod(filepath.Join(cnbDir, "config", "default.ini"), 0000))
				})
				it.After(func() {
					Expect(os.Chmod(filepath.Join(cnbDir, "config", "default.ini"), os.ModePerm)).To(Succeed())
				})
				it("returns an error", func() {
					_, _, err := files.WriteConfig(layerDir, cnbDir, phpdist.PhpIniConfig{})
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
			context("when template file is malformed", func() {
				it.Before(func() {
					Expect(os.WriteFile(filepath.Join(cnbDir, "config", "buildpack.ini"), []byte(`
include_path = "{{ .IncludePath"
{{ end }}
`), os.ModePerm)).To(Succeed())
				})
				it("returns an error", func() {
					_, _, err := files.WriteConfig(layerDir, cnbDir, phpdist.PhpIniConfig{})
					Expect(err).To(MatchError(ContainSubstring("failed to parse php.ini template")))
					Expect(err).To(MatchError(ContainSubstring("bad character")))
				})
			})
			context("when ini file can't be opened for writing", func() {
				it.Before(func() {
					Expect(os.MkdirAll(filepath.Join(layerDir, "etc"), os.ModePerm)).To(Succeed())
					Expect(os.WriteFile(filepath.Join(layerDir, "etc", "buildpack.ini"), nil, 0400)).To(Succeed())
				})
				it("returns an error", func() {
					_, _, err := files.WriteConfig(layerDir, cnbDir, phpdist.PhpIniConfig{})
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})
	context("FindExtensions", func() {
		context("when the layer directory contains exactly 1 extensions directory in the expected location", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(layerDir, "lib", "php", "extensions", "no-debug-non-zts-12345"), os.ModePerm)).To(Succeed())
			})
			it("returns the name of the directory", func() {
				dir, err := files.FindExtensions(layerDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(dir).To(Equal(filepath.Join(layerDir, "lib", "php", "extensions", "no-debug-non-zts-12345")))
			})
		})
		context("failure cases", func() {
			context("when the layer directory cannot be opened", func() {
				it("returns an error", func() {
					_, err := files.FindExtensions(`\\/\/`)
					Expect(err).To(MatchError(ContainSubstring("syntax error in pattern")))
				})
			})
			context("when no extensions directories exist", func() {
				it.Before(func() {
					Expect(os.MkdirAll(filepath.Join(layerDir, "lib", "php", "extensions", "not-an-extensions-dir"), os.ModePerm)).To(Succeed())
				})
				it("returns an error stating that zero extensions dirs were found", func() {
					_, err := files.FindExtensions(layerDir)
					Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("expected 1 PHP extensions dir matching '%s/lib/php/extensions/no-debug-non-zts-*', but found 0", layerDir))))
				})
			})
			context("when multiple extensions directories exist", func() {
				it.Before(func() {
					Expect(os.MkdirAll(filepath.Join(layerDir, "lib", "php", "extensions", "no-debug-non-zts-12345"), os.ModePerm)).To(Succeed())
					Expect(os.MkdirAll(filepath.Join(layerDir, "lib", "php", "extensions", "no-debug-non-zts-98765"), os.ModePerm)).To(Succeed())
				})
				it("returns an error stating that zero extensions dirs were found", func() {
					_, err := files.FindExtensions(layerDir)
					Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("expected 1 PHP extensions dir matching '%s/lib/php/extensions/no-debug-non-zts-*', but found 2", layerDir))))
				})
			})
		})
	})
}
