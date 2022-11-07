package main

import (
	"github.com/joshuatcasey/libdependency/retrieve"
	"github.com/joshuatcasey/libdependency/versionology"
	"github.com/paketo-buildpacks/php-dist/retrieval"
)

// main is the entrypoint for retrieving new versions
func main() {
	getAllVersionsWrapper := func() (versionology.VersionFetcherArray, error) {
		impl := retrieval.GetPhpReleasesImpl{}
		return retrieval.GetAllVersions(impl)
	}

	retrieve.NewMetadata("php", getAllVersionsWrapper, retrieval.GenerateMetadata)
}
