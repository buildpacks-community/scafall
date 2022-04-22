package internal

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"

	"github.com/AidanDelaney/scafall/pkg/util"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/sprig/v3"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
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

func AskPrompts(prompts *Prompts, vars map[string]interface{}, overrides map[string]string) error {
	for _, prompt := range prompts.Prompts {
		if overide, exists := overrides[prompt.Name]; exists {
			vars[prompt.Name] = overide
			continue
		}

		var result string
		var err error

		if prompt.Choices == nil || len(prompt.Choices) == 0 {
			var validateFunc promptui.ValidateFunc = requireId
			if prompt.Required {
				validateFunc = requireNonEmptyString
			}
			p := promptui.Prompt{
				Label:    prompt.Prompt,
				Default:  prompt.Default,
				Validate: validateFunc,
			}
			result, err = p.Run()
		} else {
			p := promptui.Select{
				Label: prompt.Prompt,
				Items: prompt.Choices,
			}
			_, result, err = p.Run()
		}
		if err == nil {
			vars[prompt.Name] = result
		}
	}
	return nil
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

func ReadPromptFile(bfs billy.Filesystem, name string) (*Prompts, error) {
	promptData, err := ReadFile(bfs, name)
	if err != nil {
		return nil, err
	}

	prompts := Prompts{}
	if _, err := toml.Decode(promptData, &prompts); err != nil {
		return nil, fmt.Errorf("%s file does not match required format: %s", name, err)
	}

	for _, prompt := range prompts.Prompts {
		if util.Contains(ReservedPromptVariables, prompt.Name) {
			return nil, fmt.Errorf("%s file contains reserved variable: %s", name, prompt.Name)
		}

		if prompt.Name == "" || prompt.Prompt == "" {
			return nil, fmt.Errorf("%s file contains prompt with missing name or prompt required field", name)
		}
	}

	return &prompts, nil
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

func Apply(bfs billy.Filesystem, vars map[string]interface{}) (billy.Filesystem, error) {
	outFs := memfs.New()

	err := util.Walk(bfs, "/", func(path string, info fs.FileInfo, err error) error {
		// Do not write the prompt file to the output project
		if isPrefixOf(path, IgnoredNames) {
			return nil
		}

		t, terr := transform(&vars, path)
		if terr != nil {
			return nil
		}
		tpath := string(t)

		// Checking, if embedded file is a folder.
		if info.IsDir() {
			// Create folders structure from embedded.
			if err := outFs.MkdirAll(tpath, 0755); err != nil {
				return err
			}
		}

		// Checking, if embedded file is not a folder.
		if !info.IsDir() {
			// Set file data.
			fileData, errReadFile := ReadFile(bfs, path)
			if errReadFile != nil {
				return errReadFile
			}

			transformed, tfErr := transform(&vars, fileData)
			if tfErr != nil {
				return fmt.Errorf("failed to subsitute variables in %s", tpath)
			}
			// Create file from embedded.
			if fileInfo, err := outFs.OpenFile(tpath, os.O_CREATE|os.O_RDWR, info.Mode()); err == nil {
				defer fileInfo.Close()
				if _, err := fileInfo.Write(transformed); err != nil {
					return err
				}
			} else {
				return err
			}
		}

		return nil
	})

	return outFs, err
}

func Copy(inFs billy.Filesystem, outFs billy.Filesystem) error {
	err := util.Walk(inFs, "/", func(path string, info fs.FileInfo, err error) error {
		// Checking, if embedded file is a folder.
		if info.IsDir() {
			// Create folders structure from embedded.
			if err := outFs.MkdirAll(path, 0755); err != nil {
				return err
			}
		}

		// Checking, if embedded file is not a folder.
		if !info.IsDir() {
			// create a copy
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
			log.Default().Printf("    %s  %s", "create", path)
		}

		return nil
	})
	return err
}

func transform(env *map[string]interface{}, data string) ([]byte, error) {
	var output bytes.Buffer
	tpl, err := template.New("bp").Funcs(sprig.FuncMap()).Parse(data)
	if err != nil {
		return nil, errors.New("cannot parse file template")
	}
	err = tpl.Execute(&output, *env)
	if err != nil {
		return nil, errors.New("cannot replace variables in file template")
	}
	return output.Bytes(), err
}
