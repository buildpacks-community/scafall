package internal_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/scafall/pkg/internal"

	h "github.com/buildpacks/pack/testhelpers"
	"github.com/sclevine/spec"
)

func testApply(t *testing.T, when spec.G, it spec.S) {
	when("Applying to a filesystem", func() {
		it("correctly replaces strings in a filesytem", func() {
			tmpDir, _ := ioutil.TempDir("", "test")
			defer os.RemoveAll(tmpDir)
			outputDir, _ := ioutil.TempDir("", "test")
			defer os.RemoveAll(outputDir)
			err := os.MkdirAll(filepath.Join(tmpDir, "/{{.Foo}}/{{.Foo}}"), 0766)
			h.AssertNil(t, err)
			f, err := os.Create(filepath.Join(tmpDir, "/{{.Foo}}/{{.Foo}}/{{.Foo}}.txt"))
			h.AssertNil(t, err)
			f.Write([]byte("{{.Foo}}"))
			f.Close()
			vars := map[string]string{"Foo": "Bar"}

			err = internal.Apply(tmpDir, vars, outputDir)
			h.AssertNil(t, err)

			bar, err := os.Open(filepath.Join(outputDir, "/Bar/Bar/Bar.txt"))
			h.AssertNil(t, err)
			h.AssertNotNil(t, bar)

			var c string
			c, err = internal.ReadFile(filepath.Join(outputDir, "/Bar/Bar/Bar.txt"))
			h.AssertNil(t, err)
			h.AssertContains(t, c, "Bar")
		})
	})
}

func testApplyNoArgument(t *testing.T, when spec.G, it spec.S) {
	when("Applying to a file without argument", func() {
		it("does not replace the template variable", func() {
			tmpDir, _ := ioutil.TempDir("", "test")
			defer os.RemoveAll(tmpDir)
			outputDir, _ := ioutil.TempDir("", "test")
			defer os.RemoveAll(outputDir)
			testFile := filepath.Join(tmpDir, "test.txt")
			content := "{{ .Foo }}"
			os.WriteFile(testFile, []byte(content), 0600)

			err := internal.Apply(tmpDir, nil, outputDir)
			h.AssertNil(t, err)

			c, err := internal.ReadFile(filepath.Join(outputDir, "test.txt"))
			h.AssertNil(t, err)
			h.AssertContains(t, c, content)
		})
	})

	when("Applying to a filesystem without argument", func() {
		it("does not replace the template variable", func() {
			tmpDir, _ := ioutil.TempDir("", "test")
			defer os.RemoveAll(tmpDir)
			outputDir, _ := ioutil.TempDir("", "test")
			defer os.RemoveAll(outputDir)
			err := os.MkdirAll(filepath.Join(tmpDir, "/{{.Foo}}/{{.Foo}}"), 0766)
			h.AssertNil(t, err)
			f, err := os.Create(filepath.Join(tmpDir, "/{{.Foo}}/{{.Foo}}/{{.Foo}}.txt"))
			h.AssertNil(t, err)
			f.Write([]byte("{{.Foo}}"))
			f.Close()
			vars := map[string]string{"Bar": "bar"}

			err = internal.Apply(tmpDir, vars, outputDir)
			h.AssertNil(t, err)

			fooTxt := filepath.Join(outputDir, "/{{.Foo}}/{{.Foo}}/{{.Foo}}.txt")
			foo, err := os.Stat(fooTxt)
			h.AssertNil(t, err)
			h.AssertNotNil(t, foo)

			var c string
			c, err = internal.ReadFile(filepath.Join(outputDir, "/{{.Foo}}/{{.Foo}}/{{.Foo}}.txt"))
			h.AssertNil(t, err)
			h.AssertContains(t, c, "{{.Foo}}")
		})
	})
}
