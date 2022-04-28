// Scafall creates new source projects from project templates.  Project
// templates are stored in git repositories and new source projects are created
// on your local filesystem.
package scafall

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AidanDelaney/scafall/pkg/internal"
	"github.com/imdario/mergo"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
)

// Scafall allows programmatic control over the default values for variables
// Overrides are skipped in prompts but can be locally overridden in a
// `.override.toml` file.
type Scafall struct {
	Overrides     map[string]string
	DefaultValues map[string]interface{}
	OutputFolder  string
}

type ScafallOption func(*Scafall)

func WithOutputFolder(folder string) ScafallOption {
	return func(s *Scafall) {
		s.OutputFolder = folder
	}
}

func WithOverrides(overrides map[string]string) ScafallOption {
	return func(s *Scafall) {
		s.Overrides = overrides
	}
}

func WithDefaultValues(defaults map[string]interface{}) ScafallOption {
	return func(s *Scafall) {
		s.DefaultValues = defaults
	}
}

// Create a new Scafall with the given options.
func NewScafall(opts ...ScafallOption) Scafall {
	var (
		defaultOverrides     = map[string]string{}
		defautlDefaultValues = map[string]interface{}{}
		defaultOutputFolder  = "."
	)

	s := Scafall{
		Overrides:     defaultOverrides,
		DefaultValues: defautlDefaultValues,
		OutputFolder:  defaultOutputFolder,
	}

	for _, opt := range opts {
		opt(&s)
	}

	return s
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
	mergedOverrides := make(map[string]string)
	mergo.Merge(&mergedOverrides, s.Overrides)
	mergo.Merge(&mergedOverrides, overrides)

	values, err := internal.AskPrompts(prompts, mergedOverrides, vars, os.Stdin)
	if err != nil {
		return err
	}
	if _, exists := values[varName]; !exists {
		return fmt.Errorf("can not process the chosen element of collection: '%s'", varName)
	}
	choice := values[varName]
	inFs, err = inFs.Chroot(choice)
	if err != nil {
		return nil
	}
	return create(s, inFs, outputDir)
}

// ScaffoldCollection creates a project after prompting the end-user to choose
// one of the projects in the collection at url.
func (s Scafall) ScaffoldCollection(url string, prompt string) error {
	inFs, err := urlToFs(url)
	if err != nil {
		return err
	}
	return collection(s, inFs, s.OutputFolder, prompt)
}

// Scaffold accepts url containing project templates and creates an output
// project.  The url can either point to a project template or a collection of
// project templates.
func (s Scafall) Scaffold(url string) error {
	inFs, err := urlToFs(url)
	if err != nil {
		return err
	}

	if isCollection(inFs) {
		return collection(s, inFs, s.OutputFolder, "Choose a project template")
	}
	return create(s, inFs, s.OutputFolder)
}

func create(s Scafall, bfs billy.Filesystem, targetDir string) error {
	var values map[string]string

	// Create prompts and merge any overrides
	if _, err := bfs.Stat(internal.PromptFile); err == nil {
		prompts, err := internal.ReadPromptFile(bfs, internal.PromptFile)
		if err != nil {
			return err
		}
		overrides := map[string]string{}
		if _, err := bfs.Stat(internal.OverrideFile); err == nil {
			overrides, err = internal.ReadOverrides(bfs, internal.OverrideFile)
			if err != nil {
				return err
			}
		}
		mergedOverrides := make(map[string]string)
		mergo.Merge(&mergedOverrides, s.Overrides)
		mergo.Merge(&mergedOverrides, overrides)
		values, err = internal.AskPrompts(prompts, mergedOverrides, s.DefaultValues, os.Stdin)
		if err != nil {
			return err
		}
		mergo.Merge(&values, mergedOverrides)
	}

	transformedFs := memfs.New()
	errApply := internal.Apply(bfs, values, transformedFs)
	if errApply != nil {
		return fmt.Errorf("failed to load new project skeleton: %s", errApply)
	}

	os.MkdirAll(targetDir, 0755)
	outFs := osfs.New(targetDir)
	errCopy := internal.Copy(transformedFs, outFs)
	if errCopy != nil {
		return fmt.Errorf("failed to write new project: %s", errCopy)
	}

	return nil
}
