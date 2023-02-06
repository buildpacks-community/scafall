package scafall_integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sclevine/spec"
	h "github.com/stretchr/testify/assert"

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
				outputDir = t.TempDir()
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
				h.Nil(t, err)
				err = s.Scaffold()
				h.Nil(t, err)

				templateFile := filepath.Join(outputDir, "template.go")
				_, err = os.Stat(templateFile)
				h.Nil(t, err)
				data, _ := os.ReadFile(templateFile)

				for _, s := range currentCase.promptAnswers {
					h.Contains(t, string(data), s)
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
			outputDir = t.TempDir()
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
			h.Nil(t, err)
			_, err = os.Stat(filepath.Join(outputDir, "prompts.toml"))
			h.NotNil(t, err)
			_, err = os.Stat(filepath.Join(outputDir, "README.txt"))
			h.NotNil(t, err)
			data, _ := os.ReadFile(templateFile)
			h.Contains(t, string(data), "QUACK")

			templateBinary := filepath.Join(outputDir, "quack", "quack.jpg")
			fi, err := os.Stat(templateBinary)
			h.Nil(t, err)
			h.NotEqual(t, 0, fi)
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
			outputDir = t.TempDir()
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
			h.Nil(t, err)
			data, _ := os.ReadFile(templateFile)
			h.Contains(t, string(data), "this is not a test")
		})

		it.After(func() {
			os.RemoveAll(outputDir)
		})
	})

	when("An invalid template is passed", func() {
		it("reports template errors and does not output a project", func() {
			brokenTemplate := "testdata/broken"
			outputDir := t.TempDir()
			defer os.RemoveAll(outputDir)

			s, _ := scafall.NewScafall(brokenTemplate, scafall.WithOutputFolder(outputDir))
			err := s.Scaffold()
			h.NotNil(t, err)

			templateFile := filepath.Join(outputDir, "template.go")
			_, err = os.Stat(templateFile)
			h.NotNil(t, err)
		})
	})

	when("various sprig functions are used", func() {
		it("parses and executes correctly", func() {
			template := "testdata/sprig_templates"
			outputDir := t.TempDir()
			defer os.RemoveAll(outputDir)

			s, _ := scafall.NewScafall(template,
				scafall.WithOutputFolder(outputDir),
				scafall.WithArguments(map[string]string{
					"TestPrompt": "quack.exe",
				}),
			)
			err := s.Scaffold()
			h.Nil(t, err)

			readmeFile := filepath.Join(outputDir, "TEMPLATES.txt")
			contents, err := os.ReadFile(readmeFile)
			h.Nil(t, err)
			text := string(contents)
			h.Contains(t, text, "* {{ .Unknown | snake_case }}")

			h.Contains(t, text, "* HELLO!")
			h.Contains(t, text, "* 1m35s")
			h.Contains(t, text, "* .exe")
		})
	})
}
