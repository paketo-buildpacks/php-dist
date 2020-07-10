package main

import (
	"github.com/paketo-buildpacks/packit"
	phpdist "github.com/paketo-buildpacks/php-dist"
)

func main() {
	packit.Run(phpdist.Detect(), phpdist.Build())
}
