package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildpack"

	"github.com/cloudfoundry/dagger"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

func TestIntegration(t *testing.T) {
	RegisterTestingT(t)
	spec.Run(t, "Integration", testIntegration, spec.Report(report.Terminal{}))
}

func testIntegration(t *testing.T, when spec.G, it spec.S) {
	when("push simple app", func() {
		it("servers simple php page", func() {
			uri, err := dagger.PackageBuildpack()
			Expect(err).ToNot(HaveOccurred())

			//set up buildermetadata for 'old way' of using pack in dagger
			builderMetadata := dagger.BuilderMetadata{
				Buildpacks: []dagger.Buildpack{
					{
						ID:  "org.cloudfoundry.buildpacks.php",
						URI: uri,
					},
				},
				Groups: []dagger.Group{
					{
						[]buildpack.Info{
							{
								ID:      "org.cloudfoundry.buildpacks.php",
								Version: "0.0.1",
							},
						},
					},
				},
			}

			app, err := dagger.Pack(filepath.Join("fixtures", "simple_app"), builderMetadata, dagger.CFLINUXFS3)
			Expect(err).ToNot(HaveOccurred())

			app.SetHealthCheck("", "3s", "1s")

			err = app.Start()
			if err != nil {
				_, err = fmt.Fprintf(os.Stderr, "App failed to start: %v\n", err)
				containerID, imageName, volumeIDs, err := app.ContainerInfo()
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("ContainerID: %s\nImage Name: %s\nAll leftover cached volumes: %v\n", containerID, imageName, volumeIDs)

				containerLogs, err := app.ContainerLogs()
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("Container Logs:\n %s\n", containerLogs)
				t.FailNow()
			}

			resp, err := app.HTTPGetSucceeds("/index.php?date")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(resp)).To(ContainSubstring("SUCCESS"))

			Expect(app.Destroy()).To(Succeed()) //Only destroy app if the test passed to leave artifacts to debug
		})

		it("runs a cli app", func() {
			uri, err := dagger.PackageBuildpack()
			Expect(err).ToNot(HaveOccurred())

			//set up buildermetadata for 'old way' of using pack in dagger
			builderMetadata := dagger.BuilderMetadata{
				Buildpacks: []dagger.Buildpack{
					{
						ID:  "org.cloudfoundry.buildpacks.php",
						URI: uri,
					},
				},
				Groups: []dagger.Group{
					{
						[]buildpack.Info{
							{
								ID:      "org.cloudfoundry.buildpacks.php",
								Version: "0.0.1",
							},
						},
					},
				},
			}

			app, err := dagger.Pack(filepath.Join("fixtures", "simple_cli_app"), builderMetadata, dagger.CFLINUXFS3)
			Expect(err).ToNot(HaveOccurred())

			app.SetHealthCheck("", "3s", "1s")

			err = app.Start()
			if err != nil {
				_, err = fmt.Fprintf(os.Stderr, "App failed to start: %v\n", err)
				containerID, imageName, volumeIDs, err := app.ContainerInfo()
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("ContainerID: %s\nImage Name: %s\nAll leftover cached volumes: %v\n", containerID, imageName, volumeIDs)

				containerLogs, err := app.ContainerLogs()
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("Container Logs:\n %s\n", containerLogs)
				t.FailNow()
			}

			logs, err := app.ContainerLogs()
			Expect(err).ToNot(HaveOccurred())
			Expect(logs).To(ContainSubstring("SUCCESS"))

			Expect(app.Destroy()).To(Succeed()) //Only destroy app if the test passed to leave artifacts to debug
		})
	})
}
