package integration_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cloudfoundry/dagger"
	"github.com/paketo-buildpacks/occam"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var phpDistBuildpack string

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	root, err := filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	bpPath := filepath.Join(root, "artifact")
	cmd := exec.Command("scripts/package.sh", "--archive")
	cmd.Env = append(os.Environ(), fmt.Sprintf("PACKAGE_DIR=%s", bpPath))
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	Expect(err).ToNot(HaveOccurred())

	phpDistBuildpack = fmt.Sprintf("%s.tgz", bpPath)
	// TODO offline

	defer func() {
		Expect(dagger.DeleteBuildpack(phpDistBuildpack)).To(Succeed())
	}()

	SetDefaultEventuallyTimeout(10 * time.Second)

	suite := spec.New("Integration", spec.Parallel(), spec.Report(report.Terminal{}))
	suite("SimpleApp", testSimpleApp)
	suite.Run(t)
}

func ContainerLogs(id string) func() string {
	docker := occam.NewDocker()

	return func() string {
		logs, _ := docker.Container.Logs.Execute(id)
		return logs.String()
	}
}

func GetGitVersion() (string, error) {
	gitExec := pexec.NewExecutable("git")
	revListOut := bytes.NewBuffer(nil)

	err := gitExec.Execute(pexec.Execution{
		Args:   []string{"rev-list", "--tags", "--max-count=1"},
		Stdout: revListOut,
	})
	if err != nil {
		return "", err
	}

	stdout := bytes.NewBuffer(nil)
	err = gitExec.Execute(pexec.Execution{
		Args:   []string{"describe", "--tags", strings.TrimSpace(revListOut.String())},
		Stdout: stdout,
	})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(strings.TrimPrefix(stdout.String(), "v")), nil
}
