// Scafall creates new source projects from project templates.  Project
// templates are stored in git repositories and new source projects are created
// on your local filesystem.
package scafall

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AidanDelaney/scafall/pkg/internal"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
)

// Scafall allows programmatic control over pre-populated Variables and control
// over the variables that are allowed to be prompted.  Reserverd variables are
// skipped in prompts.
type Scafall struct {
	Variables map[string]interface{}
	Reserved  []string
}

// Create a New Scafall with the given pre-populated Variables and reserved
// variables
func New(vars map[string]interface{}, reservedPromptValues []string) Scafall {
	return Scafall{
		Variables: vars,
		Reserved:  reservedPromptValues,
	}
}

// Present a local directory or a git repo as a Filesystem
func urlToFs(url string) (billy.Filesystem, error) {
	var inFs billy.Filesystem

	// if the URL is a local folder, then do not git clone it
	if _, err := os.Stat(url); err == nil {
		inFs = osfs.New(url)
	} else {
		inFs = memfs.New()
		_, err := git.Clone(memory.NewStorage(), inFs, &git.CloneOptions{
			URL:   url,
			Depth: 1,
		})
		if err != nil {
			return nil, err
		}
	}

	return inFs, nil
}

// If there is no top level prompts and some subdirectories contain prompts,
// then we're dealing with a collection.  Otherwise it's scaffolding with no
// prompts
func isCollection(bfs billy.Filesystem) bool {
	if _, err := bfs.Stat(internal.PromptFile); err == nil {
		return false
	}

	entries, err := bfs.ReadDir("/")
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			promptFile := filepath.Join(entry.Name(), internal.PromptFile)
			if _, err := bfs.Stat(promptFile); err == nil {
				return true
			}
		}
	}
	return false
}

func collection(s Scafall, inFs billy.Filesystem, outputDir string, prompt string) error {
	varName := "__ScaffoldUrl"
	vars := map[string]interface{}{}

	choices := []string{}
	entries, err := inFs.ReadDir("/")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != ".git" {
			choices = append(choices, entry.Name())
		}
	}

	initialPrompt := internal.Prompt{
		Name:     varName,
		Prompt:   prompt,
		Required: true,
		Choices:  choices,
	}
	prompts := internal.Prompts{
		Prompts: []internal.Prompt{initialPrompt},
	}
	overrides, err := internal.ReadOverrides(inFs, internal.OverrideFile)
	if err != nil {
		return err
	}

	internal.AskPrompts(&prompts, vars, overrides)
	if _, exists := vars[varName]; !exists {
		return fmt.Errorf("can not process the chosen element of collection: '%s'", varName)
	}
	choice := vars[varName].(string)
	inFs, err = inFs.Chroot(choice)
	if err != nil {
		return nil
	}
	return create(s, inFs, outputDir)
}

// ScaffoldCollection creates a project after prompting the end-user to choose
// one of the projects in the collection at url.
func (s Scafall) ScaffoldCollection(url string, prompt string, outputDir string) error {
	inFs, err := urlToFs(url)
	if err != nil {
		return err
	}
	return collection(s, inFs, outputDir, prompt)
}

// Scaffold accepts url containing project templates and creates an output
// project.  The url can either point to a project template or a collection of
// project templates.
func (s Scafall) Scaffold(url string, outputDir string) error {
	inFs, err := urlToFs(url)
	if err != nil {
		return err
	}

	if isCollection(inFs) {
		return collection(s, inFs, outputDir, "Choose a project template")
	}
	return create(s, inFs, outputDir)
}

func create(s Scafall, bfs billy.Filesystem, targetDir string) error {
	// don't clobber any existing files
	if _, ok := os.Stat(targetDir); ok == nil {
		return fmt.Errorf("directory %s already exists", targetDir)
	}

	var transformedFs = bfs

	if _, err := bfs.Stat(internal.PromptFile); err == nil {
		prompts, err := internal.ReadPromptFile(bfs, internal.PromptFile)
		if err != nil {
			return err
		}
		overides := map[string]string{}
		if _, err := bfs.Stat(internal.OverrideFile); err == nil {
			overides, err = internal.ReadOverrides(bfs, internal.OverrideFile)
			if err != nil {
				return err
			}
		}
		err = internal.AskPrompts(prompts, s.Variables, overides)
		if err != nil {
			return err
		}
	}

	transformedFs, err := internal.Apply(bfs, s.Variables)
	if err != nil {
		return err
	}

	os.MkdirAll(targetDir, 0755)
	outFs := osfs.New(targetDir)
	err = internal.Copy(transformedFs, outFs)
	if err != nil {
		return fmt.Errorf("failed to load new project skeleton: %s", err)
	}

	return nil
}
