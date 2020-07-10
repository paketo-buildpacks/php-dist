package main

import (
	"github.com/paketo-buildpacks/packit"
	phpdist "github.com/paketo-buildpacks/php-dist"
)

func main() {
	buildpackYMLParser := phpdist.NewBuildpackYMLParser()
	packit.Run(
		phpdist.Detect(
			buildpackYMLParser,
		),
		phpdist.Build(),
	)
}
