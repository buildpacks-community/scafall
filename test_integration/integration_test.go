package scafall_integration_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	h "github.com/buildpacks/pack/testhelpers"
	"github.com/sclevine/spec"

	scafall "github.com/buildpacks/scafall/pkg"
)

func testIntegration(t *testing.T, when spec.G, it spec.S) {
	type TestCase struct {
		title         string
		folder        []string
		promptAnswers []string
	}
	testCases := []TestCase{
		{"Test no prompt file", []string{"testdata", "empty"}, []string{}},
		{"Test empty prompt file", []string{"testdata", "noprompts"}, []string{}},
		{"Test string prompts", []string{"testdata", "str_prompts"}, []string{"test"}},
		{"Test required prompts", []string{"testdata", "requireprompts"}, []string{"test"}},
	}

	for _, testCase := range testCases {
		currentCase := testCase

		when(currentCase.title, func() {
			var (
				outputDir string
			)

			it.Before(func() {
				outputDir, _ = ioutil.TempDir("", "test")
			})

			it("creates a template file", func() {
				inputTemplate := filepath.Join(currentCase.folder...)
				if _, err := os.Stat(inputTemplate); err != nil {
					panic(fmt.Errorf("cannot open input template %s", inputTemplate))
				}

				s, err := scafall.NewScafall(
					inputTemplate,
					scafall.WithOutputFolder(outputDir),
				)
				h.AssertNil(t, err)
				err = s.Scaffold()
				h.AssertNil(t, err)

				templateFile := filepath.Join(outputDir, "template.go")
				_, err = os.Stat(templateFile)
				h.AssertNil(t, err)
				data, _ := ioutil.ReadFile(templateFile)

				for _, s := range currentCase.promptAnswers {
					h.AssertContains(t, string(data), s)
				}
			})

			it.After(func() {
				os.RemoveAll(outputDir)
			})
		})
	}

	when("A file or folder contains a template term", func() {
		var (
			outputDir string
		)

		it.Before(func() {
			outputDir, _ = ioutil.TempDir("", "test")
		})

		it("renames a templated folder and file", func() {
			s, _ := scafall.NewScafall(
				"testdata/template_folder",
				scafall.WithArguments(map[string]string{"duck": "quack", "crow": "caw"}),
				scafall.WithOutputFolder(outputDir),
			)
			s.Scaffold()

			templateFile := filepath.Join(outputDir, "quack", "quack.go")
			_, err := os.Stat(templateFile)
			h.AssertNil(t, err)
			_, err = os.Stat(filepath.Join(outputDir, "prompts.toml"))
			h.AssertNotNil(t, err)
			_, err = os.Stat(filepath.Join(outputDir, "README.txt"))
			h.AssertNotNil(t, err)
			data, _ := ioutil.ReadFile(templateFile)
			h.AssertContains(t, string(data), "QUACK")

			templateBinary := filepath.Join(outputDir, "quack", "quack.jpg")
			fi, err := os.Stat(templateBinary)
			h.AssertNil(t, err)
			h.AssertNotEq(t, 0, fi)
		})

		it.After(func() {
			os.RemoveAll(outputDir)
		})
	})

	when("A subPath is requested", func() {
		var (
			outputDir string
		)

		it.Before(func() {
			outputDir, _ = ioutil.TempDir("", "test")
		})

		it("creates a project from a subPath", func() {
			s, _ := scafall.NewScafall(
				"testdata/collection",
				scafall.WithOutputFolder(outputDir),
				scafall.WithSubPath("two"),
			)
			s.Scaffold()

			templateFile := filepath.Join(outputDir, "template.go")
			_, err := os.Stat(templateFile)
			h.AssertNil(t, err)
			data, _ := ioutil.ReadFile(templateFile)
			h.AssertContains(t, string(data), "this is not a test")
		})

		it.After(func() {
			os.RemoveAll(outputDir)
		})
	})

	when("An invalid template is passed", func() {
		it("reports template errors and does not output a project", func() {
			brokenTemplate := "testdata/broken"
			outputDir, _ := ioutil.TempDir("", "test")

			s, _ := scafall.NewScafall(brokenTemplate, scafall.WithOutputFolder(outputDir))
			err := s.Scaffold()
			h.AssertNotNil(t, err)

			templateFile := filepath.Join(outputDir, "template.go")
			_, err = os.Stat(templateFile)
			h.AssertNotNil(t, err)
		})
	})
}
