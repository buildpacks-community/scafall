package scafall_system_test

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	h "github.com/buildpacks/pack/testhelpers"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/sclevine/spec"

	scafall "github.com/AidanDelaney/scafall/pkg"
)

// walk recursively descends path, calling walkFn
// adapted from https://golang.org/src/path/filepath/path.go
func walk(fs billy.Filesystem, path string, info os.FileInfo, walkFn filepath.WalkFunc) error {
	if !info.IsDir() {
		return walkFn(path, info, nil)
	}

	names, err := fs.ReadDir(path)
	err1 := walkFn(path, info, err)
	// If err != nil, walk can't walk into this directory.
	// err1 != nil means walkFn want walk to skip this directory or stop walking.
	// Therefore, if one of err and err1 isn't nil, walk will return.
	if err != nil || err1 != nil {
		// The caller's behavior is controlled by the return value, which is decided
		// by walkFn. walkFn may ignore err and return nil.
		// If walkFn returns SkipDir, it will be handled by the caller.
		// So walk should return whatever walkFn returns.
		return err1
	}

	for _, fileInfo := range names {
		filename := filepath.Join(path, fileInfo.Name())
		fileInfo, err := fs.Lstat(filename)
		if err != nil {
			if err := walkFn(filename, fileInfo, err); err != nil && err != filepath.SkipDir {
				return err
			}
		} else {
			err = walk(fs, filename, fileInfo, walkFn)
			if err != nil {
				if !fileInfo.IsDir() || err != filepath.SkipDir {
					return err
				}
			}
		}
	}
	return nil
}

// Walk walks the file tree rooted at root, calling fn for each file or directory in the tree, including root.
//
// All errors that arise visiting files and directories are filtered by fn: see the WalkFunc documentation for details.
//
// The files are walked in lexical order, which makes the output deterministic but requires Walk to read an entire directory into memory before proceeding to walk that directory.
//
// Walk does not follow symbolic links.
//
// adapted from https://github.com/golang/go/blob/3b770f2ccb1fa6fecc22ea822a19447b10b70c5c/src/path/filepath/path.go#L500
func walkFs(fs billy.Filesystem, root string, walkFn filepath.WalkFunc) error {
	info, err := fs.Lstat(root)

	if err != nil {
		err = walkFn(root, nil, err)
	} else {
		err = walk(fs, root, info, walkFn)
	}
	if err == filepath.SkipDir {
		return nil
	}
	return err
}

func testSystem(t *testing.T, when spec.G, it spec.S) {
	when("top level command is executed", func() {
		var (
			testFolder = filepath.Join("testdata", "bash")
			expected   = memfs.New()
			outputDir  string
		)

		it.Before(func() {
			expected.MkdirAll("bin", 0755)
			expected.OpenFile("bin/build", os.O_CREATE, 0744)
			expected.OpenFile("bin/detect", os.O_CREATE, 0744)

			outputDir, _ = ioutil.TempDir("", "test")
		})

		it("scaffolds a project", func() {
			pwd, _ := os.Getwd()
			url := filepath.Join(pwd, testFolder)

			s, _ := scafall.NewScafall(url, scafall.WithOutputFolder(outputDir))
			err := s.Scaffold()
			h.AssertNil(t, err)

			bfs := osfs.New(outputDir)
			walkFs(expected, "/", func(path string, info fs.FileInfo, err error) error {
				fi, e := bfs.Stat(path)
				h.AssertNil(t, e)

				h.AssertEq(t, fi.Mode()&01000, info.Mode()&01000)
				return nil
			})
		})

		it.After(func() {
			os.RemoveAll(outputDir)
		})
	})

	when("top level command is executed", func() {
		var (
			expected  = memfs.New()
			outputDir string
		)

		it.Before(func() {
			expected.MkdirAll("quack", 0755)
			expected.OpenFile("quack/print_pi.py", os.O_CREATE, 0744)

			outputDir, _ = ioutil.TempDir("", "test")
		})

		it("scaffolds a project from a URL ", func() {
			url := "http://github.com/AidanDelaney/scafall-python-eg.git"
			arguments := map[string]string{
				"ProjectName":   "quack",
				"PythonVersion": "python3.10",
				"NumDigits":     "42",
			}

			s, _ := scafall.NewScafall(url, scafall.WithOutputFolder(outputDir), scafall.WithArguments(arguments))
			err := s.Scaffold()
			h.AssertNil(t, err)

			bfs := osfs.New(outputDir)
			walkFs(expected, "/", func(path string, info fs.FileInfo, err error) error {
				fi, e := bfs.Stat(path)
				h.AssertNil(t, e)

				h.AssertEq(t, fi.Mode()&01000, info.Mode()&01000)
				return nil
			})
		})

		it.After(func() {
			os.RemoveAll(outputDir)
		})
	})
}

func testArgs(t *testing.T, when spec.G, it spec.S) {
	when("args command is executed", func() {
		var (
			outputDir string
		)

		it.Before(func() {
			outputDir, _ = ioutil.TempDir("", "test")
		})

		it("shows arguments of a project from a URL", func() {
			url := "http://github.com/AidanDelaney/scafall-python-eg.git"

			s, _ := scafall.NewScafall(url, scafall.WithOutputFolder(outputDir))
			_, args, err := s.TemplateArguments()
			h.AssertNil(t, err)

			h.AssertEq(t, args, []string{
				"ProjectName (default: pyexample)",
				"PythonVersion=python3.10, python3.9, python3.8 (default: python3.10)",
				"NumDigits (default: 3)",
			})
		})

		it("shows arguments of a template collection", func() {
			url := "https://github.com/AidanDelaney/cnb-buildpack-templates"

			s, _ := scafall.NewScafall(url, scafall.WithOutputFolder(outputDir))
			_, args, err := s.TemplateArguments()
			h.AssertNil(t, err)

			h.AssertEq(t, args, []string{"Go", "bash"})
		})

		it.After(func() {
			os.RemoveAll(outputDir)
		})
	})
}
