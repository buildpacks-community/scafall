package internal_test

import (
	"testing"

	"github.com/AidanDelaney/scafall/pkg/internal"
	h "github.com/buildpacks/pack/testhelpers"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/sclevine/spec"
)

func testReadPrompt(t *testing.T, when spec.G, it spec.S) {
	when("Reading a prompt file", func() {
		it("reads a correct prompt file", func() {
			bfs := memfs.New()
			correctPromptFile := "[[prompt]]\nname=\"Foo\"\nprompt=\"Chhose a foo\""
			f, _ := bfs.Create(internal.PromptFile)
			f.Write([]byte(correctPromptFile))
			f.Close()

			_, err := internal.ReadPromptFile(bfs, internal.PromptFile)
			h.AssertNil(t, err)
		})

		var incorrectPromptFiles = []string{
			"incorrect",
			"[[prompt]]",
			"[[prompt]]\nname=\"test\"",
			"[[prompt]]\nprompt=\"test\"",
		}
		for _, file := range incorrectPromptFiles {
			var incorrectPromptFile = file
			when("Reading an incorrect prompt file", func() {
				bfs := memfs.New()
				it.Before(func() {
					f, _ := bfs.Create(internal.PromptFile)
					f.Write([]byte(incorrectPromptFile))
					f.Close()
				})

				it("fails with an incorrect prompt file", func() {
					_, err := internal.ReadPromptFile(bfs, internal.PromptFile)
					h.AssertNotEq(t, nil, err)
				})
			})
		}
	})
}

func testApply(t *testing.T, when spec.G, it spec.S) {
	when("Applying to a filesystem", func() {
		it("correctlyy replaces strings in a filessytem", func() {
			var bfs = memfs.New()
			err := bfs.MkdirAll("/{{.Foo}}/{{.Foo}}", 0766)
			h.AssertNil(t, err)
			f, err := bfs.Create("/{{.Foo}}/{{.Foo}}/{{.Foo}}.txt")
			h.AssertNil(t, err)
			f.Write([]byte("{{.Foo}}"))
			f.Close()
			vars := map[string]interface{}{
				"Foo": "Bar",
			}

			outFs, err := internal.Apply(bfs, vars)
			h.AssertNil(t, err)

			bar, err := outFs.Open("/Bar/Bar/Bar.txt")
			h.AssertNil(t, err)
			h.AssertNotNil(t, bar)

			var c string
			c, err = internal.ReadFile(outFs, "/Bar/Bar/Bar.txt")
			h.AssertNil(t, err)
			h.AssertContains(t, c, "Bar")
		})
	})
}
