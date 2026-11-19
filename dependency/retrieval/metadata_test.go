package main

import (
	"encoding/json"
	"testing"

	"github.com/paketo-buildpacks/libdependency/versionology"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

func TestDependencyMetadataTargets(t *testing.T) {
	targets := getSupportedPlatformStackTargets()
	dependencies := make([]versionology.Dependency, 0, len(targets))
	for _, target := range targets {
		dep, err := versionology.NewDependency(cargo.ConfigMetadataDependency{
			ID: "php", Version: "8.4.0", Stacks: target.stacks, OS: target.os, Arch: target.arch,
		}, target.target)
		if err != nil {
			t.Fatal(err)
		}
		dependencies = append(dependencies, dep)
	}

	data, err := json.Marshal(dependencies)
	if err != nil {
		t.Fatal(err)
	}
	var output []struct {
		Target string   `json:"target"`
		OS     string   `json:"os"`
		Arch   string   `json:"arch"`
		Stacks []string `json:"stacks"`
	}
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatal(err)
	}
	if len(output) != 6 {
		t.Fatalf("expected 6 platform targets, got %d", len(output))
	}
	seen := map[string]bool{}
	for _, dependency := range output {
		key := dependency.Target + "/" + dependency.OS + "/" + dependency.Arch
		if seen[key] || dependency.OS != "linux" ||
			(dependency.Arch != "amd64" && dependency.Arch != "arm64") ||
			(dependency.Target != "jammy" && dependency.Target != "noble" && dependency.Target != "resolute") ||
			len(dependency.Stacks) != 1 || dependency.Stacks[0] != "io.buildpacks.stacks."+dependency.Target {
			t.Fatalf("unexpected or duplicate metadata: %+v", dependency)
		}
		seen[key] = true
	}
}
