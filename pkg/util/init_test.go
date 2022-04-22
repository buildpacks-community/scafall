package util_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUtil(t *testing.T) {
	spec.Run(t, "Walk", testWalk, spec.Report(report.Terminal{}))
}
