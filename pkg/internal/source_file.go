package internal

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	t "github.com/coveooss/gotemplate/v3/template"
)

type SourceFile struct {
	FilePath    string
	FileContent string
	FileMode    fs.FileMode
}

func (s SourceFile) Transform(inputDir string, outputDir string, vars map[string]string) error {
	outputFile, err := s.Replace(vars)
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
		inputPath := filepath.Join(inputDir, s.FilePath)
		mvErr := os.Rename(inputPath, outputPath)
		if mvErr != nil {
			return fmt.Errorf("failed to rename %s to %s", s.FilePath, outputFile.FilePath)
		}
	} else {
		os.WriteFile(outputPath, []byte(outputFile.FileContent), outputFile.FileMode|0600)
	}
	return nil
}

func replaceUnknownVars(vars map[string]string, content string) string {
	regex := regexp.MustCompile(`{{[ \t]*\.\w+`)
	transformed := content
	for _, token := range regex.FindAllString(content, -1) {
		candidate := strings.Split(token, ".")[1]
		if _, exists := vars[candidate]; !exists {
			// replace "{{\s*.candidate" with "{&{&\s*.candidate"
			replacement := strings.Replace(token, "{{", ReplacementDelimiter, 1)
			transformed = strings.ReplaceAll(transformed, token, replacement)
		}
	}
	return transformed
}

func (s SourceFile) Replace(vars map[string]string) (SourceFile, error) {
	opts := t.DefaultOptions().
		Set(t.Overwrite, t.Sprig, t.StrictErrorCheck, t.AcceptNoValue).
		Unset(t.Razor)
	template, err := t.NewTemplate(
		"",
		vars,
		"",
		opts)
	if err != nil {
		return SourceFile{}, err
	}

	filePath := replaceUnknownVars(vars, s.FilePath)
	transformedFilePath, err := template.ProcessContent(filePath, "")
	if err != nil {
		return SourceFile{}, err
	}
	transformedFilePath = strings.ReplaceAll(transformedFilePath, ReplacementDelimiter, "{{")

	transformedFileContent := ""
	if s.FileContent != "" {
		fileContent := replaceUnknownVars(vars, s.FileContent)
		transformedFileContent, err = template.ProcessContent(fileContent, "")
		if err != nil {
			return SourceFile{}, err
		}
		transformedFileContent = strings.ReplaceAll(transformedFileContent, ReplacementDelimiter, "{{")
	}

	return SourceFile{FilePath: transformedFilePath, FileContent: transformedFileContent, FileMode: s.FileMode}, nil
}
