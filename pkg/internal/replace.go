package internal

import (
	"regexp"
	"strings"

	"github.com/coveooss/gotemplate/v3/collections"
	t "github.com/coveooss/gotemplate/v3/template"
)

func replaceUnknownVars(vars collections.IDictionary, content string) string {
	regex := regexp.MustCompile(`{{[ \t]*\.\w+`)
	transformed := content
	for _, token := range regex.FindAllString(content, -1) {
		candidate := strings.Split(token, ".")[1]
		if !vars.Has(candidate) {
			// replace "{{\s*.candidate" with "{&{&\s*.candidate"
			replacement := strings.Replace(token, "{{", ReplacementDelimiter, 1)
			transformed = strings.ReplaceAll(transformed, token, replacement)
		}
	}
	return transformed
}

func Replace(vars collections.IDictionary, file SourceFile) (SourceFile, error) {
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

	filePath := replaceUnknownVars(vars, file.FilePath)
	transformedFilePath, err := template.ProcessContent(filePath, "")
	if err != nil {
		return SourceFile{}, err
	}
	transformedFilePath = strings.ReplaceAll(transformedFilePath, ReplacementDelimiter, "{{")

	transformedFileContent := ""
	if file.FileContent != "" {
		fileContent := replaceUnknownVars(vars, file.FileContent)
		transformedFileContent, err = template.ProcessContent(fileContent, "")
		if err != nil {
			return SourceFile{}, err
		}
		transformedFileContent = strings.ReplaceAll(transformedFileContent, ReplacementDelimiter, "{{")
	}

	return SourceFile{FilePath: transformedFilePath, FileContent: transformedFileContent}, nil
}
