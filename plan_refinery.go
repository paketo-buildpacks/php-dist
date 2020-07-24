package phpdist

import (
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
)

type PlanRefinery struct{}

func NewPlanRefinery() PlanRefinery {
	return PlanRefinery{}
}

func (r PlanRefinery) BillOfMaterials(dependency postal.Dependency) packit.BuildpackPlan {
	return packit.BuildpackPlan{
		Entries: []packit.BuildpackPlanEntry{
			{
				Name:    dependency.ID,
				Version: dependency.Version,
				Metadata: map[string]interface{}{
					"licenses": []string{},
					"name":     dependency.Name,
					"sha256":   dependency.SHA256,
					"stacks":   dependency.Stacks,
					"uri":      dependency.URI,
				},
			},
		},
	}
}
