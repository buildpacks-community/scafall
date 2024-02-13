package internal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

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
			require.Nil(t, err)
			_, err = file.WriteString("{{.Test}}")
			require.Nil(t, err)
		})

		it.After(func() {
			os.RemoveAll(inputDir)
			os.RemoveAll(targetDir)
		})

		it("creates valid output", func() {
			err := internal.Create(inputDir, map[string]string{"Test": "quack"}, targetDir)
			require.Nil(t, err)

			buf, err := os.ReadFile(filepath.Join(targetDir, "test.md"))
			require.Nil(t, err)
			require.Equal(t, string(buf), "quack")
		})

		when("a prompt.toml file is present", func() {
			it.Before(func() {
				_, err := os.Create(filepath.Join(inputDir, "prompts.toml"))
				require.Nil(t, err)
			})

			it("reads prompt.toml and creates valid output", func() {
				err := internal.Create(inputDir, map[string]string{"Test": "quack"}, targetDir)
				require.Nil(t, err)

				buf, err := os.ReadFile(filepath.Join(targetDir, "test.md"))
				require.Nil(t, err)
				require.Equal(t, string(buf), "quack")
			})
		})
	})
}
