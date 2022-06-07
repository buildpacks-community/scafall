package internal_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	h "github.com/buildpacks/pack/testhelpers"
	"github.com/coveooss/gotemplate/v3/collections"
	"github.com/sclevine/spec"

	"github.com/AidanDelaney/scafall/pkg/internal"
)

type ClosingBuffer struct {
	*bytes.Buffer
}

func (ClosingBuffer) Close() error {
	return nil
}

func testAskPrompts(t *testing.T, when spec.G, it spec.S) {
	type TestCase struct {
		prompts   []internal.Prompt
		text      string
		expected  collections.IDictionary
		overrides collections.IDictionary
		defaults  map[string]interface{}
	}
	prompt := internal.Prompt{
		Name:   "Duck",
		Prompt: "Make noise",
	}
	selection := internal.Prompt{
		Name:    "Duck",
		Choices: []string{"moo", "quack", "baa"},
	}

	duckQuack := collections.CreateDictionary().Add("Duck", "quack")
	testCases := []TestCase{
		{prompts: []internal.Prompt{prompt}, text: "\n", expected: collections.CreateDictionary().Add("Duck", "")},
		{prompts: []internal.Prompt{prompt}, text: "quack\n", expected: duckQuack},
		{prompts: []internal.Prompt{prompt}, text: "quack\n", expected: duckQuack, overrides: duckQuack},
		// \x0d is Enter
		{prompts: []internal.Prompt{prompt}, text: "\x0d", expected: duckQuack, overrides: collections.CreateDictionary(), defaults: map[string]interface{}{"Duck": "quack"}},
		// \x1b\x5b\x42 is the terminal escape sequence for down arrow
		{prompts: []internal.Prompt{selection}, text: "\x0d", expected: collections.CreateDictionary().Add("Duck", "moo")},
		{prompts: []internal.Prompt{selection}, text: "\x1b\x5b\x42\x0d", expected: duckQuack},
		{prompts: []internal.Prompt{selection}, text: "\x0d", expected: collections.CreateDictionary().Add("Duck", "moo")},
		{prompts: []internal.Prompt{selection}, text: "\x1b\x5b\x42\x0d", expected: duckQuack, overrides: duckQuack},
	}

	for _, test := range testCases {
		currentCase := test
		when("When the user is prompted", func() {
			var (
				input ClosingBuffer
			)

			it.Before(func() {
				input = ClosingBuffer{bytes.NewBufferString(currentCase.text)}
			})

			it("produces valid prompt values", func() {
				prompts := internal.Prompts{currentCase.prompts}
				values, err := internal.AskPrompts(prompts, currentCase.overrides, currentCase.defaults, input)
				h.AssertNil(t, err)
				h.AssertEq(t, values, currentCase.expected)
			})
		})
	}
}

func testReadPrompt(t *testing.T, when spec.G, it spec.S) {
	when("Reading a prompt file", func() {
		it("reads a correct prompt file", func() {
			tmpDir, _ := ioutil.TempDir("", "test")
			defer os.RemoveAll(tmpDir)
			promptFile := filepath.Join(tmpDir, internal.PromptFile)
			correctPromptFile := "[[prompt]]\nname=\"Foo\"\nprompt=\"Choose a foo\""
			f, _ := os.Create(promptFile)
			f.Write([]byte(correctPromptFile))
			f.Close()

			_, err := internal.ReadPromptFile(promptFile)
			h.AssertNil(t, err)
		})

		var incorrectPromptFiles = []string{
			"incorrect",
			"[[prompt]]",
			"[[prompt]]\nname=\"test\"",
			"[[prompt]]\nprompt=\"test\"",
		}
		for _, file := range incorrectPromptFiles {
			var incorrectPromptFile = file
			when("Reading an incorrect prompt file", func() {
				tmpDir, _ := ioutil.TempDir("", "test")
				promptFile := filepath.Join(tmpDir, internal.PromptFile)
				it.Before(func() {
					f, _ := os.Create(promptFile)
					f.Write([]byte(incorrectPromptFile))
					f.Close()
				})

				it("fails with an incorrect prompt file", func() {
					_, err := internal.ReadPromptFile(promptFile)
					h.AssertNotEq(t, nil, err)
				})
			})
		}
	})
}

func testApply(t *testing.T, when spec.G, it spec.S) {
	when("Applying to a filesystem", func() {
		it("correctly replaces strings in a filesytem", func() {
			tmpDir, _ := ioutil.TempDir("", "test")
			defer os.RemoveAll(tmpDir)
			outputDir, _ := ioutil.TempDir("", "test")
			defer os.RemoveAll(outputDir)
			err := os.MkdirAll(filepath.Join(tmpDir, "/{{.Foo}}/{{.Foo}}"), 0766)
			h.AssertNil(t, err)
			f, err := os.Create(filepath.Join(tmpDir, "/{{.Foo}}/{{.Foo}}/{{.Foo}}.txt"))
			h.AssertNil(t, err)
			f.Write([]byte("{{.Foo}}"))
			f.Close()
			vars := collections.CreateDictionary().Add("Foo", "Bar")

			err = internal.Apply(tmpDir, vars, outputDir)
			h.AssertNil(t, err)

			bar, err := os.Open(filepath.Join(outputDir, "/Bar/Bar/Bar.txt"))
			h.AssertNil(t, err)
			h.AssertNotNil(t, bar)

			var c string
			c, err = internal.ReadFile(filepath.Join(outputDir, "/Bar/Bar/Bar.txt"))
			h.AssertNil(t, err)
			h.AssertContains(t, c, "Bar")
		})
	})
}
