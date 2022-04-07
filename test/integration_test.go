package scafall_integration_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	scafall "github.com/AidanDelaney/scafall/pkg"
	h "github.com/buildpacks/pack/testhelpers"
	"github.com/sclevine/spec"
)

type ClosingBuffer struct {
	*bytes.Buffer
}

func (ClosingBuffer) Close() error {
	return nil
}

func testIntegration(t *testing.T, when spec.G, it spec.S) {
	type TestCase struct {
		title         string
		folder        []string
		vars          map[string]interface{}
		promptAnswers []string
	}
	testCases := []TestCase{
		{"Test no prompt file", []string{"testdata", "empty"}, map[string]interface{}{}, []string{}},
		{"Test empty prompt file", []string{"testdata", "noprompts"}, map[string]interface{}{}, []string{}},
		{"Test string prompts", []string{"testdata", "str_prompts"}, map[string]interface{}{}, []string{"test"}},
		{"Test required prompts", []string{"testdata", "requireprompts"}, map[string]interface{}{}, []string{"test"}},
	}

	for _, testCase := range testCases {
		currentCase := testCase

		when(currentCase.title, func() {
			var (
				outputDir string
				mrc       ClosingBuffer
			)

			it.Before(func() {
				outputDir, _ = ioutil.TempDir("", "test")
				prompts := fmt.Sprintf("%s\n", strings.Join(currentCase.promptAnswers, "\n"))
				mrc = ClosingBuffer{
					bytes.NewBufferString(prompts)}
			})

			it("creates a template file", func() {
				outputProject := filepath.Join(outputDir, "test")
				inputTemplate := filepath.Join(currentCase.folder...)
				if _, err := os.Stat(inputTemplate); err != nil {
					panic(fmt.Errorf("cannot open input template %s", inputTemplate))
				}

				s := scafall.Scafall{Variables: currentCase.vars, Reserved: []string{}, Stdin: mrc}
				s.Scaffold(inputTemplate, outputProject)

				templateFile := filepath.Join(outputProject, "template.go")
				_, err := os.Stat(templateFile)
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
			pwd, _ := os.Getwd()
			os.Chdir(outputDir)
			outputProject := filepath.Join(outputDir, "test")

			s := scafall.New(map[string]interface{}{"duck": "quack"}, []string{})
			s.Scaffold(filepath.Join(pwd, "testdata/template_folder"), outputProject)

			templateFile := filepath.Join(outputProject, "quack", "quack.go")
			_, err := os.Stat(templateFile)
			h.AssertNil(t, err)
			data, _ := ioutil.ReadFile(templateFile)
			h.AssertContains(t, string(data), "QUACK")

			os.Chdir(pwd)
		})

		it.After(func() {
			os.RemoveAll(outputDir)
		})
	})

	when("An invalid template is passed", func() {

	})
}
