package internal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sclevine/spec"
	h "github.com/stretchr/testify/assert"

	"github.com/buildpacks/scafall/pkg/internal"
)

func testReplace(t *testing.T, when spec.G, it spec.S) {
	type TestCase struct {
		file         internal.SourceFile
		vars         map[string]string
		expectedName string
	}

	testCases := []TestCase{
		{
			internal.SourceFile{FilePath: "{{.Foo}}", FileContent: ""},
			map[string]string{"Foo": "Bar"},
			"Bar",
		},
		{
			internal.SourceFile{FilePath: "{{.Foo}}"},
			map[string]string{"Bar": "Bar"},
			"{{.Foo}}",
		},
	}
	for _, testCase := range testCases {
		current := testCase
		when("variable replacement is called", func() {
			it("correctly replaces tokens", func() {
				output, err := current.file.Replace(current.vars)
				h.Nil(t, err)
				h.Equal(t, output.FilePath, current.expectedName)
			})
		})
	}
}

func testTransform(t *testing.T, when spec.G, it spec.S) {
	type TestCase struct {
		file            internal.SourceFile
		vars            map[string]string
		expectedName    string
		expectedContent string
	}
	testCases := []TestCase{
		{
			internal.SourceFile{FilePath: "{{.Foo}}", FileContent: "{{.Foo}}"},
			map[string]string{"Foo": "Bar"},
			"Bar",
			"Bar",
		},
		{
			internal.SourceFile{FilePath: "{{.Foo}}"},
			map[string]string{"Bar": "Bar"},
			"{{.Foo}}",
			"",
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		when("variable replacement is called", func() {
			var (
				inputDir  string
				outputDir string
			)
			it.Before(func() {
				inputDir = t.TempDir()
				outputDir = t.TempDir()
				err := os.WriteFile(filepath.Join(inputDir, testCase.file.FilePath), []byte(testCase.file.FileContent), 0400)
				h.Nil(t, err)
			})

			it("correctly replaces tokens", func() {
				err := testCase.file.Transform(inputDir, outputDir, testCase.vars)
				h.Nil(t, err)

				contents, err := os.ReadFile(filepath.Join(outputDir, testCase.expectedName))
				h.Nil(t, err)
				h.Equal(t, string(contents), testCase.expectedContent)
			})
		})
	}
}
