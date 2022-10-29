// Scafall creates new source projects from project templates.  Project
// templates are stored in git repositories and new source projects are created
// on your local filesystem.
package scafall

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/buildpacks/scafall/pkg/internal"

	"github.com/AlecAivazis/survey/v2"
)

// Scafall allows programmatic control over the default values for variables.
// Any provided Arguments cause prompts for the same variable name to be skipped.
type Scafall struct {
	URL          string
	Arguments    map[string]string
	OutputFolder string
	SubPath      string
	CloneCache   string
}

type Option func(*Scafall)

// Set the output folder in which to create scaffold a template.
func WithOutputFolder(folder string) Option {
	return func(s *Scafall) {
		s.OutputFolder = folder
	}
}

// Set values for each variable as key-value pairs.
func WithArguments(arguments map[string]string) Option {
	return func(s *Scafall) {
		s.Arguments = arguments
	}
}

// Use a sub folder within the template repository as the source for a template.
func WithSubPath(subPath string) Option {
	return func(s *Scafall) {
		s.SubPath = subPath
	}
}

// Create a new Scafall with the given options.
func NewScafall(url string, opts ...Option) (Scafall, error) {
	var (
		defaultArguments    = map[string]string{}
		defaultOutputFolder = "."
	)

	s := Scafall{
		URL:          url,
		Arguments:    defaultArguments,
		OutputFolder: defaultOutputFolder,
	}

	for _, opt := range opts {
		opt(&s)
	}

	return s, nil
}

// Scaffold accepts url containing project templates and creates an output
// project.  The url can either point to a project template or a collection of
// project templates.
func (s Scafall) Scaffold() error {
	err := s.clone()
	if err != nil {
		s.cleanUp()
		return err
	}
	inFs := s.CloneCache
	if isCollection, choices := internal.IsCollection(inFs); isCollection {
		question := survey.Select{
			Message: "choose a project template",
			Options: choices,
		}
		response := struct {
			Template string
		}{
			Template: "",
		}
		err := survey.AskOne(&question, response, survey.WithValidator(survey.Required))
		if err != nil {
			s.cleanUp()
			return err
		}
		inFs = path.Join(s.CloneCache, response.Template)
	}

	err = internal.Create(inFs, s.Arguments, s.OutputFolder)
	if err != nil {
		s.cleanUp()
	}

	return err
}

// TemplateArguments returns a list of variable names that can be passed to the template
func (s Scafall) TemplateArguments() (string, []string, error) {
	err := s.clone()
	if err != nil {
		return "", nil, err
	}
	inFs := s.CloneCache
	if isCollection, choices := internal.IsCollection(inFs); isCollection {
		return "templates available in collection", choices, nil
	}

	promptFile := filepath.Join(inFs, internal.PromptFile)
	p, err := os.Open(promptFile)
	if err != nil {
		s.cleanUp()
		return "", nil, err
	}
	template, err := internal.NewTemplate(p, nil, nil)
	if err != nil {
		s.cleanUp()
		return "", nil, err
	}
	prompts := template.Arguments()
	argsStrings := make([]string, len(prompts))
	for i, p := range prompts {
		if len(p.Choices) == 0 {
			argsStrings[i] = fmt.Sprintf("%s (default: %s)", p.Name, p.Default)
		} else {
			cString := strings.Join(p.Choices, ", ")
			argsStrings[i] = fmt.Sprintf("%s=%s (default: %s)", p.Name, cString, p.Choices[0])
		}
	}
	return "arguments offered by template", argsStrings, nil
}

func (s *Scafall) cleanUp() {
	s.CloneCache = ""
	os.RemoveAll(s.CloneCache)
	os.RemoveAll(s.OutputFolder)
}

func (s *Scafall) clone() error {
	if s.CloneCache != "" {
		return nil
	}

	tmpDir, err := os.MkdirTemp("", "scafall")
	if err != nil {
		return err
	}

	fs, err := internal.URLToFs(s.URL, s.SubPath, tmpDir)
	if err != nil {
		return err
	}
	s.CloneCache = fs
	return nil
}
