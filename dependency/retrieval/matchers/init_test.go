package matchers_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPhpDistRetrievalMatchers(t *testing.T) {
	suite := spec.New("retrieval_matchers", spec.Report(report.Terminal{}))
	suite("PhpReleasePrettyMatcher", testPhpReleasePrettyMatcher)
	suite.Run(t)
}
