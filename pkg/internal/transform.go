package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"

	"github.com/gabriel-vasile/mimetype"

	"github.com/buildpacks/scafall/pkg/internal/util"
)

const (
	OverrideFile         string = ".override.toml"
	ReplacementDelimiter string = "{&{&"
)

var (
	IgnoredNames       = []string{PromptFile, OverrideFile}
	IgnoredDirectories = []string{".git", "node_modules"}
)

func ReadFile(path string) (string, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot read file %s", path)
	}
	return string(buf), nil
}

func ReadOverrides(overrideFile string) (map[string]string, error) {
	var overrides map[string]string
	// if no override file
	if _, err := os.Stat(overrideFile); err != nil {
		return nil, nil
	}

	overrideData, err := ReadFile(overrideFile)
	if err != nil {
		return nil, err
	}

	if _, err := toml.Decode(overrideData, &overrides); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("%s file does not match required format", overrideFile))
	}

	return overrides, nil
}

func Apply(inputDir string, vars map[string]string, outputDir string) error {
	if vars == nil {
		vars = map[string]string{}
	}
	files, err := findTransformableFiles(inputDir)
	if err != nil {
		return fmt.Errorf("failed to find files in input folder: %s %s", inputDir, err)
	}

	for _, file := range files {
		err := file.Transform(inputDir, outputDir, vars)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to transform %s", file.FilePath))
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
