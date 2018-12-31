package go_tree

import (
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

type (
	Visitor interface {
		Visit(n Node) Visitor
	}

	Node struct {
		Pos
	}

	Pos struct {
		Depth int
		FileName string
		CurrentLine string
		ParentLine string
	}

	Result struct {
		DirNum int64
		FileNum int64
	}

	visitor struct {}
)

func (v visitor) Visit(n Node) Visitor {
	if n != (Node{}) {
		if n.IsHiddenFile() {
			return nil
		}

		switch n.Type() {
		case "dir":
			fmt.Println(n.ParentLine + n.CurrentLine + filepath.Base(n.FileName))
		case "file":
			fmt.Println(n.ParentLine + n.CurrentLine + filepath.Base(n.FileName))
		}
	}
	return v
}

func Walk(v Visitor, n Node, r Result) Result {
	if n == (Node{}) {
		return r
	}

	if v = v.Visit(n); v == nil {
		return r
	}

	switch n.Type() {
	case "dir":
		r = WalkDir(v, n, r)
		r.DirNum++
		return r
	case "file":
		r.FileNum++
		return r
	}

	return r
}

func WalkDir(v Visitor, n Node, r Result) Result {
	files, err := ioutil.ReadDir(n.FileName)
	if err != nil {
		panic(err)
	}

	lastIndex := len(files) - 1
	for i, file := range files {
		var nextNode Node
		if i != lastIndex {
			var parentLine string
			if n.Depth == 0 {
				parentLine = ""
			} else {
				parentLine = "│   "
			}

			nextNode = Node{
				Pos: Pos{
					Depth: n.Depth + 1,
					FileName: filepath.Join(n.FileName, file.Name()),
					ParentLine: n.ParentLine + parentLine,
					CurrentLine: "├── ",
				},
			}
		} else {
			var parentLine string
			if n.Depth == 0 {
				parentLine = ""
			} else {
				parentLine = "│   "
			}
			nextNode = Node{
				Pos: Pos{
					Depth: n.Depth + 1,
					FileName: filepath.Join(n.FileName, file.Name()),
					ParentLine: n.ParentLine + parentLine,
					CurrentLine: "└── ",
				},
			}
		}
		r = Walk(v, nextNode, r)
	}

	return r
}

func (n Node) Type() string {
	fileInfo, err := os.Stat(n.FileName)
	if err != nil {
		panic(err)
	}

	if fileInfo.IsDir() {
		return "dir"
	} else {
		return "file"
	}
}

func (n Node) IsHiddenFile() bool {
	// TODO: Check hidden file in Windows
	// https://grokbase.com/t/gg/golang-nuts/144va1n8w5/go-nuts-how-do-check-if-file-or-directory-is-hidden-under-windows
	if runtime.GOOS != "windows" {
		baseFileName := filepath.Base(n.FileName)
		if baseFileName[0:1] == "." {
			return true
		} else {
			return false
		}
	}

	return false
}

func TreeCommand(c *cli.Context) error {
	fmt.Println(".")

	rootDir, _ := os.Getwd()

	v := visitor{}
	r := Result{}
	n := Node{
		Pos {
			Depth: 0,
			FileName: rootDir,
			ParentLine: "",
		},
	}

	r = WalkDir(v, n, r)

	fmt.Printf("\n%d directories, %d files\n", r.DirNum, r.FileNum)
	return nil
}
