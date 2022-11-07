package matchers_test

import (
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/php-dist/retrieval"
	"github.com/paketo-buildpacks/php-dist/retrieval/matchers"
	"github.com/sclevine/spec"
)

func testPhpReleasePrettyMatcher(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		now              *time.Time
		phpReleasePretty retrieval.PhpReleasePretty
	)

	it.Before(func() {
		temp := time.Now()
		now = &temp
		phpReleasePretty = retrieval.PhpReleasePretty{
			SemverVersion: semver.MustParse("1.2.3"),
			Sha256:        "abc123",
			ReleaseDate:   now,
		}
	})

	context("PhpReleasePrettyMatcher", func() {
		it("matches empty", func() {
			match, err := matchers.NewPhpReleasePrettyMatcher(retrieval.PhpReleasePretty{}).Match(retrieval.PhpReleasePretty{})
			Expect(err).NotTo(HaveOccurred())
			Expect(match).To(BeTrue())
		})

		it("matches full", func() {
			match, err := matchers.NewPhpReleasePrettyMatcher(phpReleasePretty).Match(phpReleasePretty)
			Expect(err).NotTo(HaveOccurred())
			Expect(match).To(BeTrue())
		})

		it("requires semverVersion to match", func() {
			match, err := matchers.NewPhpReleasePrettyMatcher(phpReleasePretty).Match(retrieval.PhpReleasePretty{
				Sha256:      "abc123",
				ReleaseDate: now,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(match).To(BeFalse())
		})

		it("requires sha256 to match", func() {
			match, err := matchers.NewPhpReleasePrettyMatcher(phpReleasePretty).Match(retrieval.PhpReleasePretty{
				SemverVersion: semver.MustParse("1.2.3"),
				ReleaseDate:   now,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(match).To(BeFalse())
		})

		it("requires releaseDate to match", func() {
			match, err := matchers.NewPhpReleasePrettyMatcher(phpReleasePretty).Match(retrieval.PhpReleasePretty{
				SemverVersion: semver.MustParse("1.2.3"),
				Sha256:        "abc123",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(match).To(BeFalse())
		})

		context("Failure Cases", func() {
			it("requires a PhpReleasePretty", func() {
				_, err := matchers.NewPhpReleasePrettyMatcher(retrieval.PhpReleasePretty{}).Match("")
				Expect(err).To(MatchError("PhpReleasePrettyMatcher expects an retrieval.PhpReleasePretty, received string"))
			})
		})
	})
}
