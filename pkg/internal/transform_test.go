package internal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sclevine/spec"
	h "github.com/stretchr/testify/assert"

	"github.com/buildpacks-community/scafall/pkg/internal"
)

func testApply(t *testing.T, when spec.G, it spec.S) {
	when("Applying to a filesystem", func() {
		it("correctly replaces strings in a filesytem", func() {
			tmpDir := t.TempDir()
			outputDir := t.TempDir()
			err := os.MkdirAll(filepath.Join(tmpDir, "/{{.Foo}}/{{.Foo}}"), 0766)
			h.Nil(t, err)
			f, err := os.Create(filepath.Join(tmpDir, "/{{.Foo}}/{{.Foo}}/{{.Foo}}.txt"))
			h.Nil(t, err)
			f.Write([]byte("{{.Foo}}"))
			f.Close()
			vars := map[string]string{"Foo": "Bar"}

			err = internal.Apply(tmpDir, vars, outputDir)
			h.Nil(t, err)

			bar, err := os.Open(filepath.Join(outputDir, "/Bar/Bar/Bar.txt"))
			h.Nil(t, err)
			h.NotNil(t, bar)

			var c string
			c, err = internal.ReadFile(filepath.Join(outputDir, "/Bar/Bar/Bar.txt"))
			h.Nil(t, err)
			h.Contains(t, c, "Bar")
		})
	})
}

func testApplyNoArgument(t *testing.T, when spec.G, it spec.S) {
	when("Applying to a file without argument", func() {
		it("does not replace the template variable", func() {
			tmpDir := t.TempDir()
			outputDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.txt")
			content := "{{ .Foo }}"
			os.WriteFile(testFile, []byte(content), 0600)

			err := internal.Apply(tmpDir, nil, outputDir)
			h.Nil(t, err)

			c, err := internal.ReadFile(filepath.Join(outputDir, "test.txt"))
			h.Nil(t, err)
			h.Contains(t, c, content)
		})
	})

	when("Applying to a filesystem without argument", func() {
		it("does not replace the template variable", func() {
			tmpDir := t.TempDir()
			outputDir := t.TempDir()
			err := os.MkdirAll(filepath.Join(tmpDir, "/{{.Foo}}/{{.Foo}}"), 0766)
			h.Nil(t, err)
			f, err := os.Create(filepath.Join(tmpDir, "/{{.Foo}}/{{.Foo}}/{{.Foo}}.txt"))
			h.Nil(t, err)
			f.Write([]byte("{{.Foo}}"))
			f.Close()
			vars := map[string]string{"Bar": "bar"}

			err = internal.Apply(tmpDir, vars, outputDir)
			h.Nil(t, err)

			fooTxt := filepath.Join(outputDir, "/{{.Foo}}/{{.Foo}}/{{.Foo}}.txt")
			foo, err := os.Stat(fooTxt)
			h.Nil(t, err)
			h.NotNil(t, foo)

			var c string
			c, err = internal.ReadFile(filepath.Join(outputDir, "/{{.Foo}}/{{.Foo}}/{{.Foo}}.txt"))
			h.Nil(t, err)
			h.Contains(t, c, "{{.Foo}}")
		})
	})
}
