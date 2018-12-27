package main

import (
	"fmt"
	"github.com/cloudfoundry/php-cnb/php"
	"os"

	"github.com/buildpack/libbuildpack/buildplan"

	"github.com/cloudfoundry/libcfbuildpack/build"
)

func main() {
	buildContext, err := build.DefaultBuild()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to build: %s", err)
		os.Exit(101)
	}

	code, err := runBuild(buildContext)
	if err != nil {
		buildContext.Logger.Info(err.Error())
	}

	os.Exit(code)

}

func runBuild(context build.Build) (int, error) {
	context.Logger.FirstLine(context.Logger.PrettyIdentity(context.Buildpack))

	contributor, willContribute, err := php.NewContributor(context)
	if err != nil {
		return context.Failure(102), err
	}

	if willContribute {
		err := contributor.Contribute()
		if err != nil {
			return context.Failure(103), err
		}
	}

	return context.Success(buildplan.BuildPlan{})
}
