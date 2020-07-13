package phpdist_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
	phpdist "github.com/paketo-buildpacks/php-dist"
	"github.com/paketo-buildpacks/php-dist/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir         string
		workingDir        string
		cnbDir            string
		entryResolver     *fakes.EntryResolver
		dependencyManager *fakes.DependencyManager
		clock             chronos.Clock
		timeStamp         time.Time
		planRefinery      *fakes.BuildPlanRefinery
		buffer            *bytes.Buffer

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = ioutil.TempDir("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), []byte(`api = "0.2"
[buildpack]
  id = "org.some-org.some-buildpack"
  name = "Some Buildpack"
  version = "some-version"

[metadata]
  [metadata.default-versions]
    php = "7.2.*"

  [[metadata.dependencies]]
    deprecation_date = 2021-04-01T00:00:00Z
    id = "some-dep"
    name = "Some Dep"
    sha256 = "some-sha"
    stacks = ["some-stack"]
    uri = "some-uri"
    version = "some-dep-version"
`), 0644)
		Expect(err).NotTo(HaveOccurred())

		entryResolver = &fakes.EntryResolver{}
		entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
			Name:    "php",
			Version: "7.2.*",
			Metadata: map[string]interface{}{
				"version-source": "buildpack.yml",
			},
		}

		dependencyManager = &fakes.DependencyManager{}
		dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{Name: "PHP"}

		timeStamp = time.Now()
		clock = chronos.NewClock(func() time.Time {
			return timeStamp
		})

		planRefinery = &fakes.BuildPlanRefinery{}

		timeStamp = time.Now()
		clock = chronos.NewClock(func() time.Time {
			return timeStamp
		})

		planRefinery.BillOfMaterialCall.Returns.BuildpackPlan = packit.BuildpackPlan{
			Entries: []packit.BuildpackPlanEntry{
				{
					Name:    "php",
					Version: "7.2.*",
					Metadata: map[string]interface{}{
						"version-source": "buildpack.yml",
					},
				},
			},
		}

		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		buffer = bytes.NewBuffer(nil)
		logEmitter := phpdist.NewLogEmitter(buffer)

		build = phpdist.Build(entryResolver, dependencyManager, planRefinery, logEmitter, clock)
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("returns a result that builds correctly", func() {
		result, err := build(packit.BuildContext{
			WorkingDir: workingDir,
			CNBPath:    cnbDir,
			Stack:      "some-stack",
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "some-version",
			},
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name:    "php",
						Version: "7.2.*",
						Metadata: map[string]interface{}{
							"version-source": "buildpack.yml",
						},
					},
				},
			},
			Layers: packit.Layers{Path: layersDir},
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(result).To(Equal(packit.BuildResult{
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name:    "php",
						Version: "7.2.*",
						Metadata: map[string]interface{}{
							"version-source": "buildpack.yml",
						},
					},
				},
			},
			Layers: []packit.Layer{
				{
					Name:      "php",
					Path:      filepath.Join(layersDir, "php"),
					SharedEnv: packit.Environment{},
					BuildEnv:  packit.Environment{},
					LaunchEnv: packit.Environment{},
					Build:     false,
					Launch:    true,
					Cache:     false,
					// Metadata: map[string]interface{}{
					// 	php.DepKey: "",
					// 	"built_at":        timeStamp.Format(time.RFC3339Nano),
					// },
				},
			},
		}))

		Expect(filepath.Join(layersDir, "php")).To(BeADirectory())

		Expect(entryResolver.ResolveCall.Receives.BuildpackPlanEntrySlice).To(Equal([]packit.BuildpackPlanEntry{
			{
				Name:    "php",
				Version: "7.2.*",
				Metadata: map[string]interface{}{
					"version-source": "buildpack.yml",
				},
			},
		}))

		Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
		Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("php"))
		Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal("7.2.*"))
		Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

		Expect(dependencyManager.InstallCall.Receives.Dependency).To(Equal(postal.Dependency{Name: "PHP"}))
		Expect(dependencyManager.InstallCall.Receives.CnbPath).To(Equal(cnbDir))
		Expect(dependencyManager.InstallCall.Receives.LayerPath).To(Equal(filepath.Join(layersDir, "php")))

		Expect(buffer.String()).To(ContainSubstring("Some Buildpack some-version"))
		Expect(buffer.String()).To(ContainSubstring("Resolving PHP version"))
		Expect(buffer.String()).To(ContainSubstring("Selected PHP version (using buildpack.yml): "))
		Expect(buffer.String()).To(ContainSubstring("Executing build process"))
	})

	context("when the build plan entry includes the build flag", func() {
		var workingDir string

		it.Before(func() {
			var err error
			workingDir, err = ioutil.TempDir("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
				Name:    "php",
				Version: "7.2.*",
				Metadata: map[string]interface{}{
					"version-source": "buildpack.yml",
					"build":          true,
				},
			}

			planRefinery.BillOfMaterialCall.Returns.BuildpackPlan = packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name:    "php",
						Version: "7.2.*",
						Metadata: map[string]interface{}{
							"version-source": "buildpack.yml",
							"build":          true,
						},
					},
				},
			}
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

		it("marks the php layer as cached", func() {
			result, err := build(packit.BuildContext{
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				WorkingDir: workingDir,
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name:    "php",
							Version: "7.2.*",
							Metadata: map[string]interface{}{
								"version-source": "buildpack.yml",
								"build":          true,
							},
						},
					},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.BuildResult{
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name:    "php",
							Version: "7.2.*",
							Metadata: map[string]interface{}{
								"version-source": "buildpack.yml",
								"build":          true,
							},
						},
					},
				},
				Layers: []packit.Layer{
					{
						Name:      "php",
						Path:      filepath.Join(layersDir, "php"),
						SharedEnv: packit.Environment{},
						BuildEnv:  packit.Environment{},
						LaunchEnv: packit.Environment{},
						Build:     true,
						Launch:    true,
						Cache:     true,
						// Metadata: map[string]interface{}{
						// 	phpdist.DepKey: "",
						// 	"built_at":        timeStamp.Format(time.RFC3339Nano),
						// },
					},
				},
			}))
		})
	})
}
