package scafall_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestScafall(t *testing.T) {
	spec.Run(t, "Walk", testWalk, spec.Report(report.Terminal{}))
}
