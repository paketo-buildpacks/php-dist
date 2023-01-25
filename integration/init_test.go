package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

var (
	phpDistBuildpack        string
	offlinePhpDistBuildpack string
	buildPlanBuildpack      string
	version                 string

	buildpackInfo struct {
		Buildpack struct {
			ID   string
			Name string
		}
	}

	config struct {
		BuildPlan string `json:"build-plan"`
	}
)

var builder occam.Builder

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect
	pack := occam.NewPack()

	format.MaxLength = 0

	root, err := filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	file, err := os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())

	_, err = toml.NewDecoder(file).Decode(&buildpackInfo)
	Expect(err).NotTo(HaveOccurred())
	Expect(file.Close()).To(Succeed())

	file, err = os.Open("../integration.json")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewDecoder(file).Decode(&config)).To(Succeed())
	Expect(file.Close()).To(Succeed())

	buildpackStore := occam.NewBuildpackStore()

	version = "1.2.3"

	phpDistBuildpack, err = buildpackStore.Get.
		WithVersion(version).
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	offlinePhpDistBuildpack, err = buildpackStore.Get.
		WithOfflineDependencies().
		WithVersion(version).
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	buildPlanBuildpack, err = buildpackStore.Get.
		Execute(config.BuildPlan)
	Expect(err).ToNot(HaveOccurred())

	SetDefaultEventuallyTimeout(5 * time.Second)

	builder, err = pack.Builder.Inspect.Execute()
	Expect(err).NotTo(HaveOccurred())

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("LayerReuse", testReusingLayerRebuild)
	suite("Offline", testOffline)
	suite("SimpleApp", testSimpleApp)
	suite("ExtensionsLoadable", testExtensionsLoadable)
	suite.Run(t)
}
