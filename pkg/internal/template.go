package internal

import (
	"fmt"
	"io"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

const (
	PromptFile string = "prompts.toml"
)

type Prompt struct {
	Name     string   `toml:"name" binding:"required"`
	Prompt   string   `toml:"prompt" binding:"required"`
	Required bool     `toml:"required"`
	Default  string   `toml:"default"`
	Choices  []string `toml:"choices,omitempty"`
}

type Prompts struct {
	Prompts []Prompt `toml:"prompt"`
}

type Template interface {
	Arguments() []Prompt
	Ask(...survey.AskOpt) (map[string]string, error)
}

type TemplateImpl struct {
	TPrompts   Prompts
	TQuestions []*survey.Question
	TArguments map[string]string
}

func NewQuestion(prompt Prompt) survey.Question {
	p := survey.Question{
		Name: prompt.Name,
	}
	if len(prompt.Choices) != 0 {
		sselect := survey.Select{
			Message: prompt.Prompt,
			Options: prompt.Choices,
			Default: prompt.Choices[0],
		}
		if prompt.Default != "" {
			sselect.Default = prompt.Default
		}
		p.Prompt = &sselect
	} else {
		input := survey.Input{
			Message: prompt.Prompt,
		}
		if prompt.Default != "" {
			input.Default = prompt.Default
		}
		p.Prompt = &input
	}

	if prompt.Required {
		p.Validate = survey.ComposeValidators(survey.Required)
	}
	return p
}

func NewTemplate(promptFile io.ReadCloser, arguments map[string]string) (Template, error) {
	if arguments == nil {
		arguments = map[string]string{}
	}
	prompts := Prompts{}
	if promptFile != nil {
		promptData, err := io.ReadAll(promptFile)
		if err != nil {
			return nil, err
		}

		if _, err := toml.Decode(string(promptData), &prompts); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("%s file does not match required format", promptFile))
		}
	}

	questions := make([]*survey.Question, 0)
	for _, prompt := range prompts.Prompts {
		if prompt.Name == "" || prompt.Prompt == "" {
			return nil, fmt.Errorf("%s file contains prompt with missing required field; name or prompt required", promptFile)
		}

		// Remove question from survey if an argument has been provided
		if _, ok := arguments[prompt.Name]; !ok {
			question := NewQuestion(prompt)
			questions = append(questions, &question)
		}
	}

	return TemplateImpl{
		TPrompts:   prompts,
		TQuestions: questions,
		TArguments: arguments,
	}, nil
}

func (t TemplateImpl) Arguments() []Prompt {
	return t.TPrompts.Prompts
}

func (t TemplateImpl) Ask(opts ...survey.AskOpt) (map[string]string, error) {
	response := map[string]interface{}{}
	if len(t.TQuestions) != 0 {
		err := survey.Ask(t.TQuestions, &response, opts...)
		if err != nil {
			return nil, err
		}
	}

	answers := make(map[string]string, len(response))
	for key, value := range response {
		val := ""
		core.WriteAnswer(&val, key, value)
		answers[key] = val
	}
	for key, value := range t.TArguments {
		answers[key] = value
	}
	return answers, nil
}
