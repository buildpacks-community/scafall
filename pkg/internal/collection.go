package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AidanDelaney/scafall/pkg/internal/util"
)

// If there is no top level prompts and some subdirectories contain prompts,
// then we're dealing with a collection.  Otherwise it's scaffolding with no
// prompts
func IsCollection(dir string) bool {
	promptFile := filepath.Join(dir, PromptFile)
	if _, err := os.Stat(promptFile); err == nil {
		return false
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			promptFile := filepath.Join(dir, entry.Name(), PromptFile)
			if _, err := os.Stat(promptFile); err == nil {
				return true
			}
		}
	}
	return false
}

func Collection(inputDir string, overrides map[string]string, defaultValues map[string]interface{}, targetDir string, prompt string) error {
	varName := "__ScaffoldUrl"
	vars := map[string]interface{}{}

	choices := []string{}
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != ".git" {
			choices = append(choices, entry.Name())
		}
	}

	initialPrompt := Prompt{
		Name:     varName,
		Prompt:   prompt,
		Required: true,
		Choices:  choices,
	}
	prompts := Prompts{
		Prompts: []Prompt{initialPrompt},
	}
	promptFile := filepath.Join(inputDir, OverrideFile)
	overridesDict, err := ReadOverrides(promptFile)
	if err != nil {
		return err
	}
	overridesDict = overridesDict.Merge(util.ToIDictionary(overrides))

	values, err := AskPrompts(prompts, overridesDict, vars, os.Stdin)
	if err != nil {
		return err
	}
	if !values.Has(varName) {
		return fmt.Errorf("can not process the chosen element of collection: '%s'", varName)
	}
	choice := values.Get(varName).(string)
	targetProject := filepath.Join(inputDir, choice)
	return Create(targetProject, overrides, defaultValues, targetDir)
}
