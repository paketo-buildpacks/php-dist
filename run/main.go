package main

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	phpdist "github.com/paketo-buildpacks/php-dist"
)

func main() {
	buildpackYMLParser := phpdist.NewBuildpackYMLParser()
	logEmitter := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))
	entryResolver := draft.NewPlanner()
	dependencyManager := postal.NewService(cargo.NewTransport())
	environment := phpdist.NewEnvironment()

	packit.Run(
		phpdist.Detect(
			buildpackYMLParser,
		),
		phpdist.Build(
			entryResolver,
			dependencyManager,
			phpdist.NewPHPFileManager(),
			environment,
			logEmitter,
			chronos.DefaultClock,
		),
	)
}
