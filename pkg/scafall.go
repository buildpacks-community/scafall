// Scafall creates new source projects from project templates.  Project
// templates are stored in git repositories and new source projects are created
// on your local filesystem.
package scafall

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/AidanDelaney/scafall/pkg/internal"
)

// Scafall allows programmatic control over the default values for variables
// Overrides are skipped in prompts but can be locally overridden in a
// `.override.toml` file.
type Scafall struct {
	Overrides    map[string]string
	OutputFolder string
	SubPath      string
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

func WithSubPath(subPath string) Option {
	return func(s *Scafall) {
		s.SubPath = subPath
	}
}

// Create a new Scafall with the given options.
func NewScafall(opts ...Option) Scafall {
	var (
		defaultOverrides    = map[string]string{}
		defaultOutputFolder = "."
	)

	s := Scafall{
		Overrides:    defaultOverrides,
		OutputFolder: defaultOutputFolder,
	}

	for _, opt := range opts {
		opt(&s)
	}

	return s
}

// Scaffold accepts url containing project templates and creates an output
// project.  The url can either point to a project template or a collection of
// project templates.
func (s Scafall) Scaffold(url string) error {
	tmpDir, _ := ioutil.TempDir("", "scafall")
	defer os.RemoveAll(tmpDir)

	var inFs *string
	fs, err := internal.URLToFs(url, s.SubPath, tmpDir)
	if err != nil {
		return err
	}
	inFs = &fs

	if isCollection, choices := internal.IsCollection(*inFs); isCollection {
		template, err := internal.AskQuestion("choose a project template", choices, os.Stdin)
		if err != nil {
			return err
		}
		fs = path.Join(fs, template)
		inFs = &fs
	}

	return internal.Create(*inFs, s.Overrides, s.OutputFolder)
}
