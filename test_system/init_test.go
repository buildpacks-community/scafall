package scafall_system_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestSystem(t *testing.T) {
	// Run in sequence as the tests change the pwd
	suite := spec.New("scafall integration", spec.Parallel(), spec.Report(report.Terminal{}))
	suite("scafall", testSystem)
	suite.Run(t)
}
