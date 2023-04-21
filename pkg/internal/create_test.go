package internal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sclevine/spec"
	h "github.com/stretchr/testify/assert"

	"github.com/buildpacks-community/scafall/pkg/internal"
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
			h.Nil(t, err)
			_, err = file.WriteString("{{.Test}}")
			h.Nil(t, err)
		})

		it.After(func() {
			os.RemoveAll(inputDir)
			os.RemoveAll(targetDir)
		})

		it("creates valid output", func() {
			err := internal.Create(inputDir, map[string]string{"Test": "quack"}, targetDir)
			h.Nil(t, err)

			buf, err := os.ReadFile(filepath.Join(targetDir, "test.md"))
			h.Nil(t, err)
			h.Equal(t, string(buf), "quack")
		})

		when("a prompt.toml file is present", func() {
			it.Before(func() {
				_, err := os.Create(filepath.Join(inputDir, "prompts.toml"))
				h.Nil(t, err)
			})

			it("reads prompt.toml and creates valid output", func() {
				err := internal.Create(inputDir, map[string]string{"Test": "quack"}, targetDir)
				h.Nil(t, err)

				buf, err := os.ReadFile(filepath.Join(targetDir, "test.md"))
				h.Nil(t, err)
				h.Equal(t, string(buf), "quack")
			})
		})
	})
}
