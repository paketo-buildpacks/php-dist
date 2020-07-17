package phpdist_test

import (
	"bytes"
	"testing"

	"github.com/paketo-buildpacks/packit"
	phpdist "github.com/paketo-buildpacks/php-dist"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPlanEntryResolver(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer   *bytes.Buffer
		resolver phpdist.PlanEntryResolver
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)
		resolver = phpdist.NewPlanEntryResolver(phpdist.NewLogEmitter(buffer))
	})

	context("when a buildpack.yml entry is included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name:    "php",
					Version: "composer-lock-version",
					Metadata: map[string]interface{}{
						"version-source": "composer.lock",
					},
				},
				{
					Name:    "php",
					Version: "composer-json-version",
					Metadata: map[string]interface{}{
						"version-source": "composer.json",
					},
				},
				{
					Name:    "php",
					Version: "other-version",
				},
				{
					Name:    "php",
					Version: "buildpack-yml-version",
					Metadata: map[string]interface{}{
						"version-source": "buildpack.yml",
					},
				},
				{
					Name:    "php",
					Version: "default-versions-version",
					Metadata: map[string]interface{}{
						"version-source": "default-versions",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name:    "php",
				Version: "buildpack-yml-version",
				Metadata: map[string]interface{}{
					"version-source": "buildpack.yml",
				},
			}))

			Expect(buffer.String()).To(ContainSubstring("    Candidate version sources (in priority order):"))
			Expect(buffer.String()).To(ContainSubstring("      buildpack.yml    -> \"buildpack-yml-version\""))
			Expect(buffer.String()).To(ContainSubstring("      composer.lock    -> \"composer-lock-version\""))
			Expect(buffer.String()).To(ContainSubstring("      composer.json    -> \"composer-json-version\""))
			Expect(buffer.String()).To(ContainSubstring("      default-versions -> \"default-versions-version\""))
			Expect(buffer.String()).To(ContainSubstring("      <unknown>        -> \"other-version\""))
		})
	})

	context("when a composer.lock entry is included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name:    "php",
					Version: "other-version",
				},
				{
					Name:    "php",
					Version: "default-versions-version",
					Metadata: map[string]interface{}{
						"version-source": "default-versions",
					},
				},
				{
					Name:    "php",
					Version: "composer-lock-version",
					Metadata: map[string]interface{}{
						"version-source": "composer.lock",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name:    "php",
				Version: "composer-lock-version",
				Metadata: map[string]interface{}{
					"version-source": "composer.lock",
				},
			}))
		})
	})

	context("when a composer.json entry is included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name:    "php",
					Version: "other-version",
				},
				{
					Name:    "php",
					Version: "default-versions-version",
					Metadata: map[string]interface{}{
						"version-source": "default-versions",
					},
				},
				{
					Name:    "php",
					Version: "composer-json-version",
					Metadata: map[string]interface{}{
						"version-source": "composer.json",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name:    "php",
				Version: "composer-json-version",
				Metadata: map[string]interface{}{
					"version-source": "composer.json",
				},
			}))
		})
	})

	context("when both a composer.json & composer.lock entry is included", func() {
		it("resolves to either of them", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name:    "php",
					Version: "other-version",
				},
				{
					Name:    "php",
					Version: "default-versions-version",
					Metadata: map[string]interface{}{
						"version-source": "default-versions",
					},
				},
				{
					Name:    "php",
					Version: "composer-json-version",
					Metadata: map[string]interface{}{
						"version-source": "composer.json",
					},
				},
				{
					Name:    "php",
					Version: "composer-lock-version",
					Metadata: map[string]interface{}{
						"version-source": "composer.lock",
					},
				},
			})
			Expect(entry.Version).To(ContainSubstring("composer-"))
			Expect(entry.Metadata["version-source"]).To(ContainSubstring("composer."))
		})
	})

	context("when a default-versions entry is included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name:    "php",
					Version: "other-version",
				},
				{
					Name:    "php",
					Version: "default-versions-version",
					Metadata: map[string]interface{}{
						"version-source": "default-versions",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name:    "php",
				Version: "default-versions-version",
				Metadata: map[string]interface{}{
					"version-source": "default-versions",
				},
			}))
		})
	})

	context("when entry flags differ", func() {
		context("OR's them together on best plan entry", func() {
			it("has all flags", func() {
				entry := resolver.Resolve([]packit.BuildpackPlanEntry{
					{
						Name:    "php",
						Version: "composer-lock-version",
						Metadata: map[string]interface{}{
							"version-source": "composer.lock",
						},
					},
					{
						Name:    "php",
						Version: "default-versions-version",
						Metadata: map[string]interface{}{
							"version-source": "default-versions",
							"build":          true,
							"launch":         true,
						},
					},
				})
				Expect(entry).To(Equal(packit.BuildpackPlanEntry{
					Name:    "php",
					Version: "composer-lock-version",
					Metadata: map[string]interface{}{
						"version-source": "composer.lock",
						"build":          true,
						"launch":         true,
					},
				}))
			})
		})
	})

	context("when an unknown source entry is included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name:    "php",
					Version: "other-version",
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name:     "php",
				Version:  "other-version",
				Metadata: map[string]interface{}{},
			}))
		})
	})
}
