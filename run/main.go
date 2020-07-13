package main

import (
	"os"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
	phpdist "github.com/paketo-buildpacks/php-dist"
)

func main() {
	buildpackYMLParser := phpdist.NewBuildpackYMLParser()
	logEmitter := phpdist.NewLogEmitter(os.Stdout)
	entryResolver := phpdist.NewPlanEntryResolver(logEmitter)
	dependencyManager := postal.NewService(cargo.NewTransport())
	planRefinery := phpdist.NewPlanRefinery()

	packit.Run(
		phpdist.Detect(
			buildpackYMLParser,
		),
		phpdist.Build(
			entryResolver,
			dependencyManager,
			planRefinery,
			logEmitter,
			chronos.DefaultClock,
		),
	)
}
