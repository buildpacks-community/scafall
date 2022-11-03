package internal

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
	cp "github.com/otiai10/copy"
	"github.com/pkg/errors"
)

// Present a local directory or a git repo as a Filesystem
func URLToFs(url string, subPath string, tmpDir string) (string, error) {
	// if the URL is a local folder, then do not git clone it
	if _, err := os.Stat(url); err == nil {
		cp.Copy(url, tmpDir)
	} else {
		_, err := git.PlainClone(tmpDir, false, &git.CloneOptions{
			URL:   url,
			Depth: 1,
		})
		if err != nil {
			return "", err
		}
	}

	requestedSubPath := path.Join(tmpDir, subPath)
	if _, err := os.Stat(requestedSubPath); err != nil {
		return "", fmt.Errorf("reequested subPath of template does not exist: %s", subPath)
	}
	return requestedSubPath, nil
}

// Create a new source project in targetDir
func Create(inputDir string, arguments map[string]string, targetDir string) error {
	promptFile := filepath.Join(inputDir, PromptFile)
	var template Template

	overridesFile := filepath.Join(inputDir, OverrideFile)
	overrides := map[string]string{}
	if _, err := os.Stat(overridesFile); err == nil {
		overrides, err = ReadOverrides(overridesFile)
		if err != nil {
			return err
		}
	}

	if _, ok := os.Stat(promptFile); ok == nil {
		p, err := os.Open(promptFile)
		if err != nil {
			return err
		}
		template, err = NewTemplate(p, arguments, overrides)
		if err != nil {
			return err
		}
	} else {
		var err error
		template, err = NewTemplate(nil, arguments, overrides)
		if err != nil {
			return err
		}
	}

	values, err := template.Ask()
	if err != nil {
		return errors.Wrap(err, "failed to prompt for values")
	}
	err = Apply(inputDir, values, targetDir)
	if err != nil {
		return errors.Wrap(err, "failed to scaffold new project")
	}

	return nil
}
