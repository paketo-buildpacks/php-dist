package phpdist

import "github.com/paketo-buildpacks/packit"

func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		return packit.DetectResult{}, nil
	}
}
