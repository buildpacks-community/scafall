package scafall_system_test

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	scafall "github.com/AidanDelaney/scafall/pkg"
	util "github.com/AidanDelaney/scafall/pkg/util"
	h "github.com/buildpacks/pack/testhelpers"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/sclevine/spec"
)

func testSystem(t *testing.T, when spec.G, it spec.S) {
	when("top level command is executed", func() {
		var (
			testFolder = filepath.Join("testdata", "bash")
			expected   = memfs.New()
			tempDir    string
			outputDir  string
		)

		it.Before(func() {
			expected.MkdirAll("bin", 0755)
			expected.OpenFile("bin/build", os.O_CREATE, 0744)
			expected.OpenFile("bin/detect", os.O_CREATE, 0744)

			tempDir, _ = ioutil.TempDir("", "test")
			outputDir = filepath.Join(tempDir, "test")
		})

		it("scaffolds a project", func() {
			pwd, _ := os.Getwd()
			url := filepath.Join(pwd, testFolder)

			s := scafall.Scafall{}
			err := s.Scaffold(url, outputDir)
			h.AssertNil(t, err)

			bfs := osfs.New(outputDir)
			util.Walk(expected, "/", func(path string, info fs.FileInfo, err error) error {
				fi, e := bfs.Stat(path)
				h.AssertNil(t, e)

				h.AssertEq(t, fi.Mode()&01000, info.Mode()&01000)
				return nil
			})

		})

		it.After(func() {
			os.RemoveAll(tempDir)
		})
	})

	when("top level command is executed", func() {
		var (
			testFolder = filepath.Join("testdata", "collection")
			expected   = memfs.New()
			tempDir    string
			outputDir  string
		)

		it.Before(func() {
			expected.Create("two.go")

			tempDir, _ = ioutil.TempDir("", "test")
			outputDir = filepath.Join(tempDir, "test")
		})

		it("scaffolds a collection", func() {
			pwd, _ := os.Getwd()
			url := filepath.Join(pwd, testFolder)

			s := scafall.Scafall{}
			err := s.Scaffold(url, outputDir)
			h.AssertNil(t, err)

			fileData, _ := ioutil.ReadFile(filepath.Join(outputDir, "two.go"))
			h.AssertContains(t, string(fileData), "test")
		})

		it.After(func() {
			os.RemoveAll(tempDir)
		})
	})
}
