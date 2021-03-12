package phpdist_test

import (
	"bytes"
	"errors"
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
		environment       *fakes.EnvironmentConfiguration
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

		entryResolver = &fakes.EntryResolver{}
		entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
			Name: "php",
			Metadata: map[string]interface{}{
				"version":        "7.2.*",
				"version-source": "buildpack.yml",
			},
		}

		dependencyManager = &fakes.DependencyManager{}
		dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{Name: "PHP"}

		timeStamp = time.Now()
		clock = chronos.NewClock(func() time.Time {
			return timeStamp
		})

		environment = &fakes.EnvironmentConfiguration{}
		planRefinery = &fakes.BuildPlanRefinery{}

		timeStamp = time.Now()
		clock = chronos.NewClock(func() time.Time {
			return timeStamp
		})

		planRefinery.BillOfMaterialsCall.Returns.BuildpackPlan = packit.BuildpackPlan{
			Entries: []packit.BuildpackPlanEntry{
				{
					Name: "php",
					Metadata: map[string]interface{}{
						"version": "7.2.*",
						"name":    "php-dependency-name",
						"sha256":  "php-dependency-sha",
						"stacks":  []string{"some-stack"},
						"uri":     "php-dependency-uri",
					},
				},
			},
		}

		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		buffer = bytes.NewBuffer(nil)
		logEmitter := phpdist.NewLogEmitter(buffer)

		build = phpdist.Build(entryResolver, dependencyManager, environment, planRefinery, logEmitter, clock)
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
						Name: "php",
						Metadata: map[string]interface{}{
							"version":        "7.2.*",
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
						Name: "php",
						Metadata: map[string]interface{}{
							"version": "7.2.*",
							"name":    "php-dependency-name",
							"sha256":  "php-dependency-sha",
							"stacks":  []string{"some-stack"},
							"uri":     "php-dependency-uri",
						},
					},
				},
			},
			Layers: []packit.Layer{
				{
					Name:             "php",
					Path:             filepath.Join(layersDir, "php"),
					SharedEnv:        packit.Environment{},
					BuildEnv:         packit.Environment{},
					LaunchEnv:        packit.Environment{},
					ProcessLaunchEnv: map[string]packit.Environment{},
					Build:            false,
					Launch:           false,
					Cache:            false,
					Metadata: map[string]interface{}{
						phpdist.DepKey: "",
						"built_at":     timeStamp.Format(time.RFC3339Nano),
					},
				},
			},
		}))

		Expect(filepath.Join(layersDir, "php")).To(BeADirectory())

		Expect(entryResolver.ResolveCall.Receives.BuildpackPlanEntrySlice).To(Equal([]packit.BuildpackPlanEntry{
			{
				Name: "php",
				Metadata: map[string]interface{}{
					"version":        "7.2.*",
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

		Expect(environment.ConfigureCall.CallCount).To(Equal(1))
		Expect(environment.ConfigureCall.Receives.Layer.Path).To(Equal(filepath.Join(layersDir, "php")))

		Expect(buffer.String()).To(ContainSubstring("Some Buildpack some-version"))
		Expect(buffer.String()).To(ContainSubstring("Resolving PHP version"))
		Expect(buffer.String()).To(ContainSubstring("Selected PHP version (using buildpack.yml): "))
		Expect(buffer.String()).To(ContainSubstring("Executing build process"))
	})

	context("when the build plan entry includes the build, launch flags", func() {
		var workingDir string

		it.Before(func() {
			var err error
			workingDir, err = ioutil.TempDir("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
				Name: "php",
				Metadata: map[string]interface{}{
					"version":        "7.2.*",
					"version-source": "buildpack.yml",
					"launch":         true,
					"build":          true,
				},
			}

			planRefinery.BillOfMaterialsCall.Returns.BuildpackPlan = packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "php",
						Metadata: map[string]interface{}{
							"version": "7.2.*",
							"name":    "php-dependency-name",
							"sha256":  "php-dependency-sha",
							"stacks":  []string{"some-stack"},
							"uri":     "php-dependency-uri",
						},
					},
				},
			}
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

		it("marks the php layer as build, cache and launch", func() {
			result, err := build(packit.BuildContext{
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				WorkingDir: workingDir,
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "php",
							Metadata: map[string]interface{}{
								"version":        "7.2.*",
								"version-source": "buildpack.yml",
								"launch":         true,
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
							Name: "php",
							Metadata: map[string]interface{}{
								"version": "7.2.*",
								"name":    "php-dependency-name",
								"sha256":  "php-dependency-sha",
								"stacks":  []string{"some-stack"},
								"uri":     "php-dependency-uri",
							},
						},
					},
				},
				Layers: []packit.Layer{
					{
						Name:             "php",
						Path:             filepath.Join(layersDir, "php"),
						SharedEnv:        packit.Environment{},
						BuildEnv:         packit.Environment{},
						LaunchEnv:        packit.Environment{},
						ProcessLaunchEnv: map[string]packit.Environment{},
						Build:            true,
						Launch:           true,
						Cache:            true,
						Metadata: map[string]interface{}{
							phpdist.DepKey: "",
							"built_at":     timeStamp.Format(time.RFC3339Nano),
						},
					},
				},
			}))
		})
	})

	context("when we refine the buildpack plan", func() {
		it.Before(func() {
			planRefinery.BillOfMaterialsCall.Returns.BuildpackPlan = packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "new-dep",
						Metadata: map[string]interface{}{
							"version":          "some-version",
							"some-extra-field": "an-extra-value",
						},
					},
				},
			}
		})
		it("refines the BuildpackPlan", func() {
			result, err := build(packit.BuildContext{
				CNBPath: cnbDir,
				Stack:   "some-stack",
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "php",
							Metadata: map[string]interface{}{
								"version":        "7.2.*",
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
							Name: "new-dep",
							Metadata: map[string]interface{}{
								"version":          "some-version",
								"some-extra-field": "an-extra-value",
							},
						},
					},
				},
				Layers: []packit.Layer{
					{
						Name:             "php",
						Path:             filepath.Join(layersDir, "php"),
						SharedEnv:        packit.Environment{},
						BuildEnv:         packit.Environment{},
						LaunchEnv:        packit.Environment{},
						ProcessLaunchEnv: map[string]packit.Environment{},
						Build:            false,
						Launch:           false,
						Cache:            false,
						Metadata: map[string]interface{}{
							phpdist.DepKey: "",
							"built_at":     timeStamp.Format(time.RFC3339Nano),
						},
					},
				},
			}))
		})
	})

	context("when there is a dependency cache match", func() {
		it.Before(func() {
			err := ioutil.WriteFile(filepath.Join(layersDir, "php.toml"), []byte("[metadata]\ndependency-sha = \"some-sha\"\n"), 0644)
			Expect(err).NotTo(HaveOccurred())

			dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
				Name:   "PHP",
				SHA256: "some-sha",
			}
		})

		it("exits build process early", func() {
			_, err := build(packit.BuildContext{
				CNBPath: cnbDir,
				Stack:   "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "php",
							Metadata: map[string]interface{}{
								"version":        "7.2.*",
								"version-source": "buildpack.yml",
							},
						},
					},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(planRefinery.BillOfMaterialsCall.CallCount).To(Equal(1))
			Expect(planRefinery.BillOfMaterialsCall.Receives.Dependency).To(Equal(postal.Dependency{
				Name:   "PHP",
				SHA256: "some-sha",
			}))

			Expect(dependencyManager.InstallCall.CallCount).To(Equal(0))

			Expect(buffer.String()).To(ContainSubstring("Some Buildpack some-version"))
			Expect(buffer.String()).To(ContainSubstring("Resolving PHP version"))
			Expect(buffer.String()).To(ContainSubstring("Selected PHP version (using buildpack.yml): "))
			Expect(buffer.String()).To(ContainSubstring("Reusing cached layer"))
			Expect(buffer.String()).ToNot(ContainSubstring("Executing build process"))
		})
	})

	context("failure cases", func() {
		context("when a dependency cannot be resolved", func() {
			it.Before(func() {
				dependencyManager.ResolveCall.Returns.Error = errors.New("failed to resolve dependency")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name: "php",
								Metadata: map[string]interface{}{
									"version":        "7.2.*",
									"version-source": "buildpack.yml",
								},
							},
						},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError("failed to resolve dependency"))
			})
		})

		context("when a dependency cannot be installed", func() {
			it.Before(func() {
				dependencyManager.InstallCall.Returns.Error = errors.New("failed to install dependency")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name: "php",
								Metadata: map[string]interface{}{
									"version":        "7.2.*",
									"version-source": "buildpack.yml",
								},
							},
						},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError("failed to install dependency"))
			})
		})

		context("when the layers directory cannot be written to", func() {
			it.Before(func() {
				Expect(os.Chmod(layersDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(layersDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name: "php",
								Metadata: map[string]interface{}{
									"version":        "7.2.*",
									"version-source": "buildpack.yml",
								},
							},
						},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})
	})
}
