package internal

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/AidanDelaney/scafall/pkg/util"
	"github.com/gabriel-vasile/mimetype"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/sprig/v3"
	"github.com/go-git/go-billy/v5"
	"github.com/manifoldco/promptui"
)

const (
	PromptFile   string = "prompts.toml"
	OverrideFile string = ".override.toml"
)

var (
	ReservedPromptVariables = []string{}
	IgnoredNames            = []string{"/" + PromptFile, "/" + OverrideFile, "/.git"}
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

func requireId(s string) error {
	return nil
}

func PreparePrompt(prompt Prompt, defaults map[string]interface{}, input io.ReadCloser) (promptui.Prompt, error) {
	var validateFunc promptui.ValidateFunc = requireId
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

func AskPrompts(prompts Prompts, overrides map[string]string, defaults map[string]interface{}, input io.ReadCloser) (map[string]string, error) {
	values := map[string]string{}
	for _, prompt := range prompts.Prompts {
		if o, exists := overrides[prompt.Name]; exists {
			values[prompt.Name] = o
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
			values[prompt.Name] = result
		}
	}
	return values, nil
}

func ReadFile(bfs billy.Filesystem, name string) (string, error) {
	file, err := bfs.Open(name)
	if err != nil {
		return "", fmt.Errorf("cannot open file %s", name)
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("cannot read file %s", name)
	}
	return string(buf), nil
}

func ReadPromptFile(bfs billy.Filesystem, name string) (Prompts, error) {
	prompts := Prompts{}
	promptData, err := ReadFile(bfs, name)
	if err != nil {
		return prompts, err
	}

	if _, err := toml.Decode(promptData, &prompts); err != nil {
		return prompts, fmt.Errorf("%s file does not match required format: %s", name, err)
	}

	for _, prompt := range prompts.Prompts {
		if util.Contains(ReservedPromptVariables, prompt.Name) {
			return prompts, fmt.Errorf("%s file contains reserved variable: %s", name, prompt.Name)
		}

		if prompt.Name == "" || prompt.Prompt == "" {
			return prompts, fmt.Errorf("%s file contains prompt with missing name or prompt required field", name)
		}
	}

	return prompts, nil
}

func ReadOverrides(bfs billy.Filesystem, name string) (map[string]string, error) {
	overrides := map[string]string{}
	// if no override file
	if _, err := bfs.Stat(name); err != nil {
		return overrides, nil
	}

	overrideData, err := ReadFile(bfs, name)
	if err != nil {
		return nil, err
	}

	if _, err := toml.Decode(overrideData, &overrides); err != nil {
		return nil, fmt.Errorf("%s file does not match required format: %s", name, err)
	}

	for k, _ := range overrides {
		if util.Contains(ReservedPromptVariables, k) {
			return nil, fmt.Errorf("%s file contains reserved variable: %s", name, k)
		}
	}

	return overrides, nil
}

func isPrefixOf(path string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func Apply(bfs billy.Filesystem, vars map[string]string, outFs billy.Filesystem) error {
	err := util.Walk(bfs, "/", func(path string, info fs.FileInfo, err error) error {
		// Do not write the prompt file to the output project
		if isPrefixOf(path, IgnoredNames) {
			return nil
		}

		tpath := path
		if t, terr := transform(vars, path); terr == nil {
			tpath = string(t)
		}

		if info.IsDir() {
			if err := outFs.MkdirAll(tpath, 0755); err != nil {
				return err
			}
		}

		if !info.IsDir() {
			if !isTextfile(bfs, path) {
				return copyBinaryFile(bfs, path, info, outFs, tpath)
			}

			return copyTextFile(bfs, path, info, vars, outFs, tpath)
		}

		return nil
	})

	return err
}

func Copy(inFs billy.Filesystem, outFs billy.Filesystem) error {
	err := util.Walk(inFs, "/", func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			if err := outFs.MkdirAll(path, 0755); err != nil {
				return err
			}
		}

		if !info.IsDir() {
			outFile, errCreateFile := outFs.OpenFile(path, os.O_CREATE|os.O_RDWR, info.Mode())
			if errCreateFile != nil {
				return fmt.Errorf("failed to create file: %s %s", path, err)
			}
			defer outFile.Close()

			inFile, errOpen := inFs.Open(path)
			if errOpen != nil {
				return fmt.Errorf("failed to copy file: %s %s", path, err)
			}
			defer inFile.Close()

			if n, errCopy := io.Copy(outFile, inFile); errCopy != nil {
				return fmt.Errorf("failed to write data to file: %s %v (%d bytes)", path, err, n)
			}
			return nil
		}
		return nil
	})

	return err
}

func isTextfile(bfs billy.Filesystem, path string) bool {
	fd, err := bfs.Open(path)
	if err != nil {
		return false
	}
	mtype, err := mimetype.DetectReader(fd)
	if err != nil {
		return false
	}

	if strings.HasPrefix(mtype.String(), "text") {
		return true
	}

	return false
}

func copyBinaryFile(bfs billy.Filesystem, path string, info fs.FileInfo, dst billy.Filesystem, dstPath string) error {
	outFile, err := dst.OpenFile(dstPath, os.O_CREATE|os.O_RDWR, info.Mode())
	if err != nil {
		return err
	}
	inFile, err := bfs.Open(path)
	if err != nil {
		return err
	}
	if n, err := io.Copy(outFile, inFile); err != nil {
		return fmt.Errorf("failed to write date to file: %s %s (%d bytes)", path, err, n)
	}
	return nil
}

func copyTextFile(bfs billy.Filesystem, path string, info fs.FileInfo, vars map[string]string, dst billy.Filesystem, dstPath string) error {
	fileData, errReadFile := ReadFile(bfs, path)
	if errReadFile != nil {
		return errReadFile
	}

	transformed, tfErr := transform(vars, fileData)
	if tfErr != nil {
		return fmt.Errorf("failed to subsitute variables in %s", path)
	}
	if fileInfo, err := dst.OpenFile(dstPath, os.O_CREATE|os.O_RDWR, info.Mode()); err == nil {
		defer fileInfo.Close()
		if _, err := fileInfo.Write(transformed); err != nil {
			return err
		}
	} else {
		return err
	}

	return nil
}

func transform(env map[string]string, data string) ([]byte, error) {
	var output bytes.Buffer
	tpl, err := template.New("bp").Funcs(sprig.FuncMap()).Parse(data)
	if err != nil {
		return nil, errors.New("cannot parse file template")
	}
	err = tpl.Execute(&output, env)
	if err != nil {
		return nil, errors.New("cannot replace variables in file template")
	}
	return output.Bytes(), err
}
