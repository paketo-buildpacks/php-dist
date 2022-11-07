package retrieval_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPhpDistRetrieval(t *testing.T) {
	suite := spec.New("retrieval", spec.Report(report.Terminal{}))
	suite("Retrieval", testRetrieval)
	suite.Run(t)
}
