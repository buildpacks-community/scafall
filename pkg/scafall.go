// Scafall creates new source projects from project templates.  Project
// templates are stored in git repositories and new source projects are created
// on your local filesystem.
package scafall

import (
	"io/ioutil"
	"os"

	"github.com/AidanDelaney/scafall/pkg/internal"
)

// Scafall allows programmatic control over the default values for variables
// Overrides are skipped in prompts but can be locally overridden in a
// `.override.toml` file.
type Scafall struct {
	Overrides     map[string]string
	DefaultValues map[string]interface{}
	OutputFolder  string
}

type Option func(*Scafall)

func WithOutputFolder(folder string) Option {
	return func(s *Scafall) {
		s.OutputFolder = folder
	}
}

func WithOverrides(overrides map[string]string) Option {
	return func(s *Scafall) {
		s.Overrides = overrides
	}
}

func WithDefaultValues(defaults map[string]interface{}) Option {
	return func(s *Scafall) {
		s.DefaultValues = defaults
	}
}

// Create a new Scafall with the given options.
func NewScafall(opts ...Option) Scafall {
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

// ScaffoldCollection creates a project after prompting the end-user to choose
// one of the projects in the collection at url.
func (s Scafall) ScaffoldCollection(url string, prompt string) error {
	tmpDir, _ := ioutil.TempDir("", "scafall")
	defer os.RemoveAll(tmpDir)

	inFs, err := internal.URLToFs(url, tmpDir)
	if err != nil {
		return err
	}
	return internal.Collection(inFs, s.Overrides, s.DefaultValues, s.OutputFolder, prompt)
}

// Scaffold accepts url containing project templates and creates an output
// project.  The url can either point to a project template or a collection of
// project templates.
func (s Scafall) Scaffold(url string) error {
	tmpDir, _ := ioutil.TempDir("", "scafall")
	defer os.RemoveAll(tmpDir)

	inFs, err := internal.URLToFs(url, tmpDir)
	if err != nil {
		return err
	}

	if internal.IsCollection(inFs) {
		return internal.Collection(inFs, s.Overrides, s.DefaultValues, s.OutputFolder, "Choose a project template")
	}
	return internal.Create(inFs, s.Overrides, s.DefaultValues, s.OutputFolder)
}
