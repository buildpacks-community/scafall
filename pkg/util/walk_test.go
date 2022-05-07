package util_test

import (
	"errors"
	"io/fs"
	"testing"

	h "github.com/buildpacks/pack/testhelpers"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/sclevine/spec"
	"golang.org/x/sys/unix"

	u "github.com/AidanDelaney/scafall/pkg/util"
)

func testWalk(t *testing.T, when spec.G, it spec.S) {
	var (
		bfs     billy.Filesystem
		listing []string
	)

	it.Before(func() {
		bfs = memfs.New()
		bfs.MkdirAll("toplevel-1/level1-1", 0711)
		bfs.Create(".hidden")
		bfs.Create("toplevel-1/afile")
		bfs.Create("toplevel-1/level1-1/file")
		util.WriteFile(bfs, "executable", []byte{}, 0744)
		listing = []string{"/", "/.hidden", "/executable", "/toplevel-1", "/toplevel-1/afile", "/toplevel-1/level1-1", "/toplevel-1/level1-1/file"}
	})
	it.After(func() {})

	when("A directory tree is walked", func() {
		it("finds all files and directories", func() {
			var found = []string{}
			u.Walk(bfs, "/", func(path string, info fs.FileInfo, err error) error {
				found = append(found, path)
				return nil
			})

			h.AssertEq(t, found, listing)
		})

		it("finds executable file permissions", func() {
			u.Walk(bfs, "/executable", func(path string, info fs.FileInfo, err error) error {
				if err := unix.Access(path, unix.X_OK); err == nil {
					return nil
				}
				return errors.New("cannot find executable file")
			})
		})
	})
}
