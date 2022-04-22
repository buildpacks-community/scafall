package scafall_integration_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestIntegration(t *testing.T) {
	// Run in sequence as the tests change the pwd
	suite := spec.New("scafall integration", spec.Sequential(), spec.Report(report.Terminal{}))
	suite("scafall", testIntegration)
	suite.Run(t)
}
