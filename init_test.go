package phpdist_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPhpDist(t *testing.T) {
	suite := spec.New("php-dist", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Build", testBuild)
	suite("Detect", testDetect)
	suite.Run(t)
}
