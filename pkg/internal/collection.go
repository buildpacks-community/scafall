package internal

import (
	"os"
	"path/filepath"
)

// If there is no top level prompts and some subdirectories contain prompts,
// then we're dealing with a collection.  Otherwise it's scaffolding with no
// prompts
func IsCollection(dir string) (bool, []string) {
	promptFile := filepath.Join(dir, PromptFile)
	if _, err := os.Stat(promptFile); err == nil {
		return false, []string{}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, []string{}
	}

	choices := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			promptFile := filepath.Join(dir, entry.Name(), PromptFile)
			if _, err := os.Stat(promptFile); err == nil {
				choices = append(choices, entry.Name())
			}
		}
	}
	return len(choices) > 0, choices
}
