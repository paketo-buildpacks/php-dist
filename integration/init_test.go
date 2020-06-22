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

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	phpDistBuildpack string
	offlinePhpDistBuildpack string
	version          string
	buildpackInfo struct {
		Buildpack struct {
			ID   string
			Name string
		}
	}
)

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	root, err := filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	file, err := os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())
	defer file.Close()

	_, err = toml.DecodeReader(file, &buildpackInfo)
	Expect(err).NotTo(HaveOccurred())

	version, err = GetGitVersion()
	Expect(err).NotTo(HaveOccurred())

	phpDistBuildpack, err = Package(root, version, false)
	Expect(err).ToNot(HaveOccurred())

	offlinePhpDistBuildpack, err = Package(root, version, true)
	Expect(err).ToNot(HaveOccurred())

	defer func() {
		Expect(os.RemoveAll(phpDistBuildpack)).To(Succeed())
		Expect(os.RemoveAll(offlinePhpDistBuildpack)).To(Succeed())
	}()

	SetDefaultEventuallyTimeout(10 * time.Second)

	suite := spec.New("Integration", spec.Parallel(), spec.Report(report.Terminal{}))
	suite("Offline", testOffline)
	suite("SimpleApp", testSimpleApp)
	suite.Run(t)
}

func Package(root, version string, cached bool) (string, error) {
	var cmd *exec.Cmd

	bpPath := filepath.Join(root, "artifact")
	if cached {
		cmd = exec.Command("scripts/package.sh", "--archive", "--cached", "--version", version)
	} else {
		cmd = exec.Command("scripts/package.sh", "--archive", "--version", version)
	}

	cmd.Env = append(os.Environ(), fmt.Sprintf("PACKAGE_DIR=%s", bpPath))
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if cached {
		return fmt.Sprintf("%s-cached.tgz", bpPath), err
	}
	return fmt.Sprintf("%s.tgz", bpPath), err
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
