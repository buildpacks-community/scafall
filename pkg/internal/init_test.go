package internal_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestInternal(t *testing.T) {
	spec.Run(t, "ReadPrompt", testReadPrompt, spec.Report(report.Terminal{}))
	spec.Run(t, "AskPrompts", testAskPrompts, spec.Report(report.Terminal{}))
}
