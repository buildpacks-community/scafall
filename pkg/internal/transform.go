package internal

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/coveooss/gotemplate/v3/collections"

	"github.com/gabriel-vasile/mimetype"
	"github.com/manifoldco/promptui"

	"github.com/AidanDelaney/scafall/pkg/internal/util"
)

const (
	PromptFile           string = "prompts.toml"
	OverrideFile         string = ".override.toml"
	ReplacementDelimiter string = "{&{&"
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

func PreparePrompt(prompt Prompt, input io.ReadCloser) (promptui.Prompt, error) {
	var validateFunc promptui.ValidateFunc = requireID
	var defaultValue = prompt.Default
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

func PrepareChoices(prompt Prompt, input io.ReadCloser) (promptui.Select, error) {
	var choices = prompt.Choices
	p := promptui.Select{
		Label: prompt.Prompt,
		Items: choices,
		Stdin: input,
	}
	return p, nil
}

func AskQuestion(question string, choices []string, input io.ReadCloser) (string, error) {
	prompt := promptui.Select{
		Label: question,
		Items: choices,
		Stdin: input,
	}
	_, result, err := prompt.Run()
	return result, err
}

func AskPrompts(prompts Prompts, arguments collections.IDictionary, input io.ReadCloser) (collections.IDictionary, error) {
	if arguments == nil {
		arguments = collections.CreateDictionary()
	}
	values := collections.CreateDictionary()
	for _, prompt := range prompts.Prompts {
		if arguments.Has(prompt.Name) {
			values.Set(prompt.Name, arguments.Get(prompt.Name))
			continue
		}

		var result string
		var err error

		if prompt.Choices == nil || len(prompt.Choices) == 0 {
			p, prepErr := PreparePrompt(prompt, input)
			if prepErr != nil {
				return nil, prepErr
			}
			result, err = p.Run()
		} else {
			p, prepErr := PrepareChoices(prompt, input)
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
	buf, err := os.ReadFile(path)
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

type SourceFile struct {
	FilePath    string
	FileContent string
	FileMode    fs.FileMode
}

func Apply(inputDir string, vars collections.IDictionary, outputDir string) error {
	files, err := findTransformableFiles(inputDir)
	if err != nil {
		return fmt.Errorf("failed to find files in input folder: %s %s", inputDir, err)
	}

	for _, file := range files {
		outputFile, err := Replace(vars, file)
		if err != nil {
			return err
		}

		dstDir := filepath.Join(outputDir, filepath.Dir(outputFile.FilePath))
		mkdirErr := os.MkdirAll(dstDir, 0744)
		if mkdirErr != nil {
			return fmt.Errorf("failed to create target directory %s", dstDir)
		}

		outputPath := filepath.Join(outputDir, outputFile.FilePath)
		if outputFile.FileContent == "" {
			inputPath := filepath.Join(inputDir, file.FilePath)
			mvErr := os.Rename(inputPath, outputPath)
			if mvErr != nil {
				return fmt.Errorf("failed to rename %s to %s", file.FilePath, outputFile.FilePath)
			}
		} else {
			os.WriteFile(outputPath, []byte(outputFile.FileContent), outputFile.FileMode|0600)
		}
	}

	return err
}

func findTransformableFiles(dir string) ([]SourceFile, error) {
	files := []SourceFile{}
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

			relPath := strings.TrimPrefix(path, dir+"/")
			if isTextfile(path) {
				fileContent, err := ReadFile(path)
				if err != nil {
					return err
				}
				fileMode := info.Type().Perm()
				files = append(files, SourceFile{FilePath: relPath, FileContent: fileContent, FileMode: fileMode})
			} else {
				files = append(files, SourceFile{FilePath: relPath, FileContent: ""})
			}
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
