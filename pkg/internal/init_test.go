package internal_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestIternal(t *testing.T) {
	spec.Run(t, "Create", testCreate, spec.Report(report.Terminal{}))
	spec.Run(t, "ReadPrompt", testReadPrompt, spec.Report(report.Terminal{}))
	spec.Run(t, "Apply", testApply, spec.Report(report.Terminal{}))
	spec.Run(t, "AskPrompts", testAskPrompts, spec.Report(report.Terminal{}))
	spec.Run(t, "NoArgument", testApplyNoArgument, spec.Report(report.Terminal{}))
	spec.Run(t, "Replace", testReplace, spec.Report(report.Terminal{}))
}
