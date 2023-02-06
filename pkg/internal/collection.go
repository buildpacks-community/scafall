package internal

import (
	"os"
	"path/filepath"
)

// If there are no top level prompts and some subdirectories contain prompts,
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

	options := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			promptFile := filepath.Join(dir, entry.Name(), PromptFile)
			if _, err := os.Stat(promptFile); err == nil {
				options = append(options, entry.Name())
			}
		}
	}
	return len(options) > 0, options
}
