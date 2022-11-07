package retrieval_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/joshuatcasey/libdependency/versionology"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/php-dist/retrieval"
	"github.com/paketo-buildpacks/php-dist/retrieval/fakes"
	"github.com/paketo-buildpacks/php-dist/retrieval/matchers"
	"github.com/sclevine/spec"
)

func testRetrieval(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		raw = retrieval.PhpReleaseAnouncementsRaw{
			"7": retrieval.PhpReleaseLineAnnouncementsRaw{
				"7.4": retrieval.PhpReleaseAnnouncementRaw{
					Date: "03 Nov 2022",
					Source: []retrieval.PhpReleaseSourceRaw{
						{
							Filename: "php-7.4.33.tar.gz",
							Sha256:   "5a2337996f07c8a097e03d46263b5c98d2c8e355227756351421003bea8f463e",
						},
						{
							Filename: "php-7.4.33.tar.bz2",
							Sha256:   "4e8117458fe5a475bf203128726b71bcbba61c42ad463dffadee5667a198a98a",
						},
						{
							Filename: "php-7.4.33.tar.xz",
							Sha256:   "924846abf93bc613815c55dd3f5809377813ac62a9ec4eb3778675b82a27b927",
						},
					},
					Version: "7.4.33",
				},
			},
			"8": retrieval.PhpReleaseLineAnnouncementsRaw{
				"8.0": retrieval.PhpReleaseAnnouncementRaw{
					Date: "27 Oct 2022",
					Source: []retrieval.PhpReleaseSourceRaw{
						{
							Filename: "php-8.0.25.tar.gz",
							Sha256:   "349a2b5a01bfccbc9af8afdf183e57bed3349706a084f3c4694aa4c7ff7cb2e9",
						},
						{
							Filename: "php-8.0.25.tar.bz2",
							Sha256:   "09d716bceb5b3db76d9023b10c1681ebbe040e51f4c18dfd35f9ff8b73bbcf8c",
						},
						{
							Filename: "php-8.0.25.tar.xz",
							Sha256:   "a291b71d0498707fc5514eb5b9513e88f0f1d4890bcdefd67282ded8a2bfb941",
						},
					},
					Version: "8.0.25",
				},
				"8.1": retrieval.PhpReleaseAnnouncementRaw{
					Date: "27 Oct 2022",
					Source: []retrieval.PhpReleaseSourceRaw{
						{
							Filename: "php-8.1.12.tar.gz",
							Sha256:   "e0e7c823c9f9aa4c021f5e34ae1a7acafc2a9f3056ca60eb70a8af8f33da3fdf",
						},
						{
							Filename: "php-8.1.12.tar.bz2",
							Sha256:   "f87d73e917facf78de7bcde53fc2faa4d4dbe0487a9406e1ab68c8ae8f33eb03",
						},
						{
							Filename: "php-8.1.12.tar.xz",
							Sha256:   "08243359e2204d842082269eedc15f08d2eca726d0e65b93fb11f4bfc51bbbab",
						},
					},
					Version: "8.1.12",
				},
			},
		}
	)

	context("GetPhpReleasesImpl", func() {
		var (
			server    *httptest.Server
			serverURL *url.URL
		)

		it.Before(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method != http.MethodGet {
					http.Error(w, "NotFound", http.StatusNotFound)
					return
				}

				jsonFilepath := filepath.Join("testdata", "php-index-2022-11-07.json")
				jsonBytes, err := os.ReadFile(jsonFilepath)
				if err != nil {
					http.Error(w, fmt.Sprintf("cannot find file %s", jsonFilepath), http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, string(jsonBytes))
			}))

			var err error
			serverURL, err = serverURL.Parse(server.URL)
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			server.Close()
		})

		it("will translate JSON into struct", func() {
			impl := retrieval.GetPhpReleasesImpl{}
			Expect(impl.Get(serverURL.String())).To(Equal(raw))
		})
	})

	context("GetAllVersions", func() {
		var getPhpReleases *fakes.GetPhpReleases

		it.Before(func() {
			getPhpReleases = &fakes.GetPhpReleases{}

			getPhpReleases.GetCall.Returns.PhpReleaseAnouncementsRaw = raw
		})

		it("will translate into VersionFetchers", func() {
			versions, err := retrieval.GetAllVersions(getPhpReleases)
			Expect(err).NotTo(HaveOccurred())
			Expect(getPhpReleases.GetCall.Receives.String).To(Equal("https://www.php.net/releases/active"))

			date20221027, err := time.Parse("2006-01-02", "2022-10-27")
			Expect(err).NotTo(HaveOccurred())

			date20221103, err := time.Parse("2006-01-02", "2022-11-03")
			Expect(err).NotTo(HaveOccurred())

			Expect(versions).To(ConsistOf(
				matchers.NewPhpReleasePrettyMatcher(retrieval.PhpReleasePretty{
					SemverVersion: semver.MustParse("7.4.33"),
					Sha256:        "5a2337996f07c8a097e03d46263b5c98d2c8e355227756351421003bea8f463e",
					ReleaseDate:   &date20221103,
				}),
				matchers.NewPhpReleasePrettyMatcher(retrieval.PhpReleasePretty{
					SemverVersion: semver.MustParse("8.0.25"),
					Sha256:        "349a2b5a01bfccbc9af8afdf183e57bed3349706a084f3c4694aa4c7ff7cb2e9",
					ReleaseDate:   &date20221027,
				}),
				matchers.NewPhpReleasePrettyMatcher(retrieval.PhpReleasePretty{
					SemverVersion: semver.MustParse("8.1.12"),
					Sha256:        "e0e7c823c9f9aa4c021f5e34ae1a7acafc2a9f3056ca60eb70a8af8f33da3fdf",
					ReleaseDate:   &date20221027,
				}),
			))
		})

		context("Failure Cases", func() {
			context("when PHP 7 release line not found", func() {
				it.Before(func() {
					delete(raw, "7")
				})

				it("will fail", func() {
					_, err := retrieval.GetAllVersions(getPhpReleases)
					Expect(err).To(MatchError("unable to find releases for PHP 7.4"))
				})
			})

			context("when PHP 7.4 release not found", func() {
				it.Before(func() {
					delete(raw["7"], "7.4")
				})

				it("will fail", func() {
					_, err := retrieval.GetAllVersions(getPhpReleases)
					Expect(err).To(MatchError("unable to find releases for PHP 7.4"))
				})
			})

			context("when PHP 8 release line not found", func() {
				it.Before(func() {
					delete(raw, "8")
				})

				it("will fail", func() {
					_, err := retrieval.GetAllVersions(getPhpReleases)
					Expect(err).To(MatchError("unable to find releases for PHP 8.0 or 8.1"))
				})
			})

			context("when PHP 8.0 release not found", func() {
				it.Before(func() {
					delete(raw["8"], "8.0")
				})

				it("will fail", func() {
					_, err := retrieval.GetAllVersions(getPhpReleases)
					Expect(err).To(MatchError("unable to find releases for PHP 8.0 or 8.1"))
				})
			})

			context("when PHP 8.1 release not found", func() {
				it.Before(func() {
					delete(raw["8"], "8.1")
				})

				it("will fail", func() {
					_, err := retrieval.GetAllVersions(getPhpReleases)
					Expect(err).To(MatchError("unable to find releases for PHP 8.0 or 8.1"))
				})
			})

			context("unable to prettify a release", func() {
				context("when the version is not semver", func() {
					it.Before(func() {
						temp := raw["8"]["8.1"]
						temp.Version = "hello"
						raw["8"]["8.1"] = temp
					})

					it("will fail", func() {
						_, err := retrieval.GetAllVersions(getPhpReleases)
						Expect(err).To(MatchError("unable to parse semantic version hello: Invalid Semantic Version"))
					})
				})

				context("when the date is not the anticipated format", func() {
					it.Before(func() {
						temp := raw["8"]["8.1"]
						temp.Date = "1999-10-16"
						raw["8"]["8.1"] = temp
					})

					it("will fail", func() {
						_, err := retrieval.GetAllVersions(getPhpReleases)
						Expect(err).To(MatchError(ContainSubstring(`unable to parse date for PHP version 8.1.12: parsing time "1999-10-16" as "02 Jan 2006":`)))
					})
				})

				context("when the Source does not include a .tar.gz file", func() {
					it.Before(func() {
						raw["8"]["8.1"].Source[0].Filename = "hello.txt"
					})

					it("will fail", func() {
						_, err := retrieval.GetAllVersions(getPhpReleases)
						Expect(err).To(MatchError("unable to find SHA256 for PHP version 8.1.12"))
					})
				})

				context("when the Source .tar.gz file has an empty sha256", func() {
					it.Before(func() {
						raw["8"]["8.1"].Source[0].Sha256 = ""
					})

					it("will fail", func() {
						_, err := retrieval.GetAllVersions(getPhpReleases)
						Expect(err).To(MatchError("unable to find SHA256 for PHP version 8.1.12"))
					})
				})
			})
		})
	})

	context("GenerateMetadata", func() {
		it("generates metadata", func() {
			releaseDate, err := time.Parse("2006-01-02", "2022-10-27")
			Expect(err).NotTo(HaveOccurred())

			phpReleasePretty := retrieval.PhpReleasePretty{
				ReleaseDate:   &releaseDate,
				SemverVersion: semver.MustParse("8.1.12"),
				Sha256:        "e0e7c823c9f9aa4c021f5e34ae1a7acafc2a9f3056ca60eb70a8af8f33da3fdf",
			}

			metadata, err := retrieval.GenerateMetadata(phpReleasePretty)
			Expect(err).NotTo(HaveOccurred())

			expectedDeprecationDate, err := time.Parse("2006-01-02", "2025-10-27")
			Expect(err).NotTo(HaveOccurred())

			Expect(metadata).To(ConsistOf(versionology.Dependency{
				ConfigMetadataDependency: cargo.ConfigMetadataDependency{
					CPE:             "cpe:2.3:a:php:php:8.1.12:*:*:*:*:*:*:*",
					DeprecationDate: &expectedDeprecationDate,
					PURL:            "pkg:generic/php@8.1.12?checksum=e0e7c823c9f9aa4c021f5e34ae1a7acafc2a9f3056ca60eb70a8af8f33da3fdf&download_url=https://www.php.net/distributions/php-8.1.12.tar.gz",
					ID:              "php",
					Licenses: []interface{}{
						"PHP-3.0",
						"PHP-3.01",
					},
					Name:           "PHP",
					Source:         "https://www.php.net/distributions/php-8.1.12.tar.gz",
					SourceChecksum: "sha256:e0e7c823c9f9aa4c021f5e34ae1a7acafc2a9f3056ca60eb70a8af8f33da3fdf",
					Stacks:         []string{"io.buildpacks.stacks.bionic"},
					Version:        "8.1.12",
				},
				SemverVersion: phpReleasePretty.Version(),
				Target:        "bionic",
			}))
		})

		context("Failure Cases", func() {
			context("when given other than PhpReleasePretty", func() {
				it("will fail", func() {
					_, err := retrieval.GenerateMetadata(versionology.NewSimpleVersionFetcher(semver.MustParse("1.2.3")))
					Expect(err).To(MatchError("expected a PhpReleasePretty"))
				})
			})
		})
	})
}
