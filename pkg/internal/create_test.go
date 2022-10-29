package internal_test

import (
	"os"
	"path/filepath"
	"testing"

	h "github.com/buildpacks/pack/testhelpers"
	"github.com/sclevine/spec"

	"github.com/buildpacks/scafall/pkg/internal"
)

func testCreate(t *testing.T, when spec.G, it spec.S) {
	when("provided with valid input", func() {
		var (
			inputDir  string
			targetDir string
		)

		it.Before(func() {
			inputDir, _ = os.MkdirTemp("", "test")
			targetDir, _ = os.MkdirTemp("", "test")
			file, err := os.Create(filepath.Join(inputDir, "test.md"))
			h.AssertNil(t, err)
			_, err = file.WriteString("{{.Test}}")
			h.AssertNil(t, err)
		})

		it.After(func() {
			os.RemoveAll(inputDir)
			os.RemoveAll(targetDir)
		})

		it("creates valid output", func() {
			err := internal.Create(inputDir, map[string]string{"Test": "quack"}, targetDir)
			h.AssertNil(t, err)

			buf, err := os.ReadFile(filepath.Join(targetDir, "test.md"))
			h.AssertNil(t, err)
			h.AssertEq(t, string(buf), "quack")
		})
	})
}
