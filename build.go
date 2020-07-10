package phpdist

import "github.com/paketo-buildpacks/packit"

func Build() packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		return packit.BuildResult{}, nil
	}
}
