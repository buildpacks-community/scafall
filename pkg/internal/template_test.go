package internal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/scafall/pkg/internal"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/Netflix/go-expect"
	pseudotty "github.com/creack/pty"
	"github.com/hinshun/vt10x"
	"github.com/sclevine/spec"
	h "github.com/stretchr/testify/assert"
)

func testReadPrompt(t *testing.T, when spec.G, it spec.S) {
	when("Reading a prompt file", func() {
		it("reads a correct prompt file", func() {
			tmpDir := t.TempDir()
			promptFile := filepath.Join(tmpDir, internal.PromptFile)
			correctPromptFile := "[[prompt]]\nname=\"Foo\"\nprompt=\"Choose a foo\""
			os.WriteFile(promptFile, []byte(correctPromptFile), 0600)

			f, err := os.Open(promptFile)
			h.Nil(t, err)
			template, err := internal.NewTemplate(f, nil, nil)
			h.Nil(t, err)
			h.Equal(t, len(template.Arguments()), 1)
		})

		it("reads incorrect prompt files", func() {
			var incorrectPromptFiles = []string{
				"incorrect",
				"[[prompt]]",
				"[[prompt]]\nname=\"test\"",
				"[[prompt]]\nprompt=\"test\"",
			}
			for _, file := range incorrectPromptFiles {
				var incorrectPromptFile = file
				tmpDir := t.TempDir()
				promptFile := filepath.Join(tmpDir, internal.PromptFile)
				os.WriteFile(promptFile, []byte(incorrectPromptFile), 0600)

				f, err := os.Open(promptFile)
				h.Nil(t, err)
				template, err := internal.NewTemplate(f, nil, nil)
				h.NotNil(t, err)
				h.Nil(t, template)
			}
		})
	})
}

type expectConsole interface {
	ExpectString(string)
	ExpectEOF()
	SendLine(string)
	Send(string)
}

type consoleWithErrorHandling struct {
	console *expect.Console
	t       *testing.T
}

func (c *consoleWithErrorHandling) ExpectString(s string) {
	if _, err := c.console.ExpectString(s); err != nil {
		c.t.Helper()
		c.t.Fatalf("ExpectString(%q) = %v", s, err)
	}
}

func (c *consoleWithErrorHandling) SendLine(s string) {
	if _, err := c.console.SendLine(s); err != nil {
		c.t.Helper()
		c.t.Fatalf("SendLine(%q) = %v", s, err)
	}
}

func (c *consoleWithErrorHandling) Send(s string) {
	if _, err := c.console.Send(s); err != nil {
		c.t.Helper()
		c.t.Fatalf("Send(%q) = %v", s, err)
	}
}

func (c *consoleWithErrorHandling) ExpectEOF() {
	if _, err := c.console.ExpectEOF(); err != nil {
		c.t.Helper()
		c.t.Fatalf("ExpectEOF() = %v", err)
	}
}

func RunTest(t *testing.T, procedure func(expectConsole), test func(terminal.Stdio) (map[string]string, error), expected map[string]string) {
	t.Helper()
	t.Parallel()

	pty, tty, err := pseudotty.Open()
	if err != nil {
		t.Fatalf("failed to open pseudotty: %v", err)
	}

	term := vt10x.New(vt10x.WithWriter(tty))
	c, err := expect.NewConsole(expect.WithStdin(pty), expect.WithStdout(term), expect.WithCloser(pty, tty))
	if err != nil {
		t.Fatalf("failed to create console: %v", err)
	}
	defer c.Close()

	donec := make(chan struct{})
	go func() {
		defer close(donec)
		procedure(&consoleWithErrorHandling{console: c, t: t})
	}()

	stdio := terminal.Stdio{In: c.Tty(), Out: c.Tty(), Err: c.Tty()}
	values, err := test(stdio)
	if err != nil {
		t.Error(err)
	}
	h.Equal(t, values, expected)

	if err := c.Tty().Close(); err != nil {
		t.Errorf("error closing Tty: %v", err)
	}
	<-donec
}

func testAskPrompts(t *testing.T, when spec.G, it spec.S) {
	type TestCase struct {
		prompts   []internal.Prompt
		text      func(c expectConsole)
		expected  map[string]string
		arguments map[string]string
	}
	prompt := internal.Prompt{
		Name:   "Duck",
		Prompt: "Make noise",
	}
	selection := internal.Prompt{
		Name:    "Duck",
		Prompt:  "Make noise",
		Choices: []string{"moo", "quack", "baa"},
	}

	duckQuack := map[string]string{"Duck": "quack"}
	testCases := []TestCase{
		{
			prompts: []internal.Prompt{prompt},
			text: func(c expectConsole) {
				c.ExpectString("Make noise")
				c.SendLine("")
				c.ExpectEOF()
			},
			expected: map[string]string{"Duck": ""}},
		{
			prompts: []internal.Prompt{prompt},
			text: func(c expectConsole) {
				c.ExpectString("Make noise")
				c.SendLine("quack")
				c.ExpectEOF()
			},
			expected: duckQuack,
		},
		{
			prompts: []internal.Prompt{prompt},
			text: func(c expectConsole) {
				c.SendLine("")
				c.ExpectEOF()
			},
			expected:  duckQuack,
			arguments: duckQuack,
		},
		// \x0d is Enter
		{
			prompts: []internal.Prompt{prompt},
			text: func(c expectConsole) {
				c.SendLine("\x0d")
				c.ExpectEOF()
			},
			expected:  duckQuack,
			arguments: duckQuack,
		},
		{
			prompts: []internal.Prompt{selection},
			text: func(c expectConsole) {
				c.ExpectString("Make noise")
				c.SendLine("\x0d")
				c.ExpectEOF()
			},
			expected: map[string]string{"Duck": "moo"},
		},
		// \x1b\x5b\x42 is the terminal escape sequence for down arrow
		{
			prompts: []internal.Prompt{selection},
			text: func(c expectConsole) {
				c.ExpectString("Make noise")
				c.SendLine("\x1b\x5b\x42\x0d")
				c.ExpectEOF()
			},
			expected: duckQuack,
		},
		{
			prompts: []internal.Prompt{selection},
			text: func(c expectConsole) {
				c.SendLine("")
				c.ExpectEOF()
			},
			expected:  duckQuack,
			arguments: duckQuack,
		},
	}

	for _, test := range testCases {
		currentCase := test
		when("When the user is prompted", func() {
			it("produces valid prompt values", func() {
				questions := []*survey.Question{}
				for _, p := range currentCase.prompts {
					q := internal.NewQuestion(p)
					questions = append(questions, &q)
				}
				prompts := internal.Prompts{Prompts: currentCase.prompts}
				template := internal.TemplateImpl{
					TPrompts:   prompts,
					TQuestions: questions,
					TArguments: currentCase.arguments,
				}

				test := func(stdio terminal.Stdio) (map[string]string, error) {
					return template.Ask(survey.WithStdio(stdio.In, stdio.Out, stdio.Err))
				}
				RunTest(t, currentCase.text, test, currentCase.expected)
			})
		})
	}
}
