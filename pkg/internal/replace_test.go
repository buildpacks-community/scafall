package internal_test

import (
	"testing"

	h "github.com/buildpacks/pack/testhelpers"
	"github.com/coveooss/gotemplate/v3/collections"
	"github.com/sclevine/spec"

	"github.com/AidanDelaney/scafall/pkg/internal"
)

func testReplace(t *testing.T, when spec.G, it spec.S) {
	type TestCase struct {
		file         internal.SourceFile
		vars         collections.IDictionary
		expectedName string
	}

	testCases := []TestCase{
		{
			internal.SourceFile{FilePath: "{{.Foo}}", FileContent: ""},
			collections.CreateDictionary().Add("Foo", "Bar"),
			"Bar",
		},
		{
			internal.SourceFile{FilePath: "{{.Foo}}"},
			collections.CreateDictionary().Add("Bar", "Bar"),
			"{{.Foo}}",
		},
	}
	for _, testCase := range testCases {
		current := testCase
		when("variable replacement is called", func() {
			it("correctly replaces tokens", func() {
				output, err := internal.Replace(current.vars, current.file)
				h.AssertNil(t, err)
				h.AssertEq(t, output.FilePath, current.expectedName)
			})
		})
	}
}
