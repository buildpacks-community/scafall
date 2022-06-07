package internal

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/sprig/v3"
	"github.com/coveooss/gotemplate/v3/collections"
	t "github.com/coveooss/gotemplate/v3/template"
	cp "github.com/otiai10/copy"

	"github.com/gabriel-vasile/mimetype"
	"github.com/manifoldco/promptui"

	"github.com/AidanDelaney/scafall/pkg/internal/util"
)

const (
	PromptFile   string = "prompts.toml"
	OverrideFile string = ".override.toml"
)

var (
	ReservedPromptVariables = []string{}
	IgnoredNames            = []string{PromptFile, OverrideFile}
	IgnoredDirectories      = []string{".git", "node_modules"}
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

func requireNonEmptyString(s string) error {
	if s == "" {
		return errors.New("please provide a non-empty value")
	}
	return nil
}

func requireID(s string) error {
	return nil
}

func PreparePrompt(prompt Prompt, defaults map[string]interface{}, input io.ReadCloser) (promptui.Prompt, error) {
	var validateFunc promptui.ValidateFunc = requireID
	var defaultValue = prompt.Default
	if k, exists := defaults[prompt.Name]; exists {
		var ok bool
		defaultValue, ok = k.(string)
		if !ok {
			return promptui.Prompt{}, fmt.Errorf("prompt for %s contains invalid string %v", prompt.Name, prompt.Default)
		}
	}
	if prompt.Required {
		validateFunc = requireNonEmptyString
	}
	p := promptui.Prompt{
		Label:    prompt.Prompt,
		Default:  defaultValue,
		Validate: validateFunc,
		Stdin:    input,
	}
	return p, nil
}

func PrepareChoices(prompt Prompt, defaults map[string]interface{}, input io.ReadCloser) (promptui.Select, error) {
	var choices = prompt.Choices
	if k, exists := defaults[prompt.Name]; exists {
		var ok bool
		choices, ok = k.([]string)
		if !ok {
			return promptui.Select{}, fmt.Errorf("prompt for %s contains invalid []string %v", prompt.Name, prompt.Default)
		}
	}
	p := promptui.Select{
		Label: prompt.Prompt,
		Items: choices,
		Stdin: input,
	}
	return p, nil
}

func AskPrompts(prompts Prompts, overrides collections.IDictionary, defaults map[string]interface{}, input io.ReadCloser) (collections.IDictionary, error) {
	if overrides == nil {
		overrides = collections.CreateDictionary()
	}
	values := collections.CreateDictionary()
	for _, prompt := range prompts.Prompts {
		if overrides.Has(prompt.Name) {
			values.Set(prompt.Name, overrides.Get(prompt.Name))
			continue
		}

		var result string
		var err error

		if prompt.Choices == nil || len(prompt.Choices) == 0 {
			p, prepErr := PreparePrompt(prompt, defaults, input)
			if prepErr != nil {
				return nil, prepErr
			}
			result, err = p.Run()
		} else {
			p, prepErr := PrepareChoices(prompt, defaults, input)
			if prepErr != nil {
				return nil, prepErr
			}
			_, result, err = p.Run()
		}
		if err == nil {
			values.Set(prompt.Name, result)
		}
	}
	return values, nil
}

func ReadFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("cannot open file %s", path)
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("cannot read file %s", path)
	}
	return string(buf), nil
}

func ReadPromptFile(promptFile string) (Prompts, error) {
	prompts := Prompts{}
	promptData, err := ReadFile(promptFile)
	if err != nil {
		return prompts, err
	}

	if _, err := toml.Decode(promptData, &prompts); err != nil {
		return prompts, fmt.Errorf("%s file does not match required format: %s", promptFile, err)
	}

	for _, prompt := range prompts.Prompts {
		if util.Contains(ReservedPromptVariables, prompt.Name) {
			return prompts, fmt.Errorf("%s file contains reserved variable: %s", promptFile, prompt.Name)
		}

		if prompt.Name == "" || prompt.Prompt == "" {
			return prompts, fmt.Errorf("%s file contains prompt with missing name or prompt required field", promptFile)
		}
	}

	return prompts, nil
}

func ReadOverrides(overrideFile string) (collections.IDictionary, error) {
	var overrides map[string]string
	// if no override file
	if _, err := os.Stat(overrideFile); err != nil {
		return collections.CreateDictionary(), nil
	}

	overrideData, err := ReadFile(overrideFile)
	if err != nil {
		return nil, err
	}

	if _, err := toml.Decode(overrideData, &overrides); err != nil {
		return nil, fmt.Errorf("%s file does not match required format: %s", overrideFile, err)
	}

	oDict := collections.CreateDictionary()
	for k, v := range overrides {
		if util.Contains(ReservedPromptVariables, k) {
			return nil, fmt.Errorf("%s file contains reserved variable: %s", overrideFile, k)
		}
		oDict.Add(k, v)
	}

	return oDict, nil
}

func Apply(inputDir string, vars collections.IDictionary, outputDir string) error {
	transformedDir, _ := ioutil.TempDir("", "scafall")
	defer os.RemoveAll(transformedDir)
	files, err := findTransformableFiles(inputDir)
	if err != nil {
		return fmt.Errorf("failed to find files in input folder: %s %s", inputDir, err)
	}

	opts := t.DefaultOptions().
		Set(t.Overwrite, t.Sprig, t.StrictErrorCheck).
		Unset(t.Razor)
	template, err := t.NewTemplate(
		transformedDir,
		vars,
		"",
		opts)
	if err != nil {
		return err
	}

	for _, file := range files {
		// replace vars in file
		filePath, _ := filepath.Rel(inputDir, file)
		transformedFilePath, err := replace(vars, filePath)
		if err != nil {
			return fmt.Errorf("failed to replace variable in filename: %s", file)
		}
		dstPath := filepath.Join(transformedDir, transformedFilePath)
		dstDir := filepath.Dir(dstPath)
		mkdirErr := os.MkdirAll(dstDir, 0744)
		if mkdirErr != nil {
			return fmt.Errorf("failed to create target directory %s", dstDir)
		}
		mvErr := os.Rename(file, dstPath)
		if mvErr != nil {
			return fmt.Errorf("failed to rename %s to %s", filePath, transformedFilePath)
		}
	}

	absoluteFilepaths, err := findTextFiles(transformedDir)
	if err != nil {
		return err
	}
	_, err = template.ProcessTemplates(transformedDir, transformedDir, absoluteFilepaths...)
	if err != nil {
		return err
	}

	err = cp.Copy(transformedDir, outputDir)
	if err != nil {
		os.RemoveAll(outputDir)
		return err
	}
	return err
}

func replace(env collections.IDictionary, data string) (string, error) {
	var output bytes.Buffer
	tpl, err := template.New("bp").Funcs(sprig.FuncMap()).Parse(data)
	if err != nil {
		return "", errors.New("cannot parse file template")
	}
	err = tpl.Execute(&output, env)
	if err != nil {
		return "", errors.New("cannot replace variables in file template")
	}
	return output.String(), err
}

func findTransformableFiles(dir string) ([]string, error) {
	files := []string{}
	err := filepath.WalkDir(dir, func(path string, info os.DirEntry, err error) error {
		if info.IsDir() && util.Contains(IgnoredDirectories, info.Name()) {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			// Ignore all prompts.toml files and any top-level README.md
			rootReadme := filepath.Join(dir, "README")
			if util.Contains(IgnoredNames, info.Name()) || strings.HasPrefix(path, rootReadme) {
				return nil
			}

			files = append(files, path)
		}
		return nil
	})

	return files, err
}

func isTextfile(path string) bool {
	fd, err := os.Open(path)
	if err != nil {
		return false
	}
	mtype, err := mimetype.DetectReader(fd)
	if err != nil {
		return false
	}

	return strings.HasPrefix(mtype.String(), "text")
}

func findTextFiles(dir string) ([]string, error) {
	files := []string{}
	err := filepath.WalkDir(dir, func(path string, info os.DirEntry, err error) error {
		if !info.IsDir() {
			if isTextfile(path) && !util.Contains(IgnoredNames, info.Name()) {
				files = append(files, path)
			}
		}
		return nil
	})

	return files, err
}
