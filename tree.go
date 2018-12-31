package go_tree

import (
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
		ParentLine []string
	}

	Result struct {
		DirNum int64
		FileNum int64
	}

	visitor struct {}
)

const (
	ThreeWayLine = "├── "
	RightAngleLine = "└── "
)

func (v visitor) Visit(n Node) Visitor {
	if n.IsHiddenFile() {
		return nil
	}

	switch n.Type() {
	case "dir":
		// TODO: coloring print
		fmt.Println(n.CurrentLine + filepath.Base(n.FileName))
	case "file":
		// TODO: coloring print
		fmt.Println(n.CurrentLine + filepath.Base(n.FileName))
	}
	return v
}

func Walk(v Visitor, n Node, r Result) Result {
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
		nextNode := n.NextNode(i, lastIndex, file.Name())
		r = Walk(v, nextNode, r)
	}

	return r
}

func (n Node) NextNode(currentIndex, lastIndex int, fileName string) Node {
	var parentLine []string
	currentLine := strings.Join(n.ParentLine, "")
	if currentIndex != lastIndex {
		if n.Depth != 0 {
			parentLine = append(n.ParentLine, "    │")
			currentLine = currentLine + "    " + ThreeWayLine
		} else {
			currentLine = ThreeWayLine
		}
	} else {
		if n.Depth != 0 {
			parentLine = append(n.ParentLine, "    ")
			currentLine = currentLine + "    " + RightAngleLine
		} else {
			currentLine = RightAngleLine
		}
	}

	return Node{
		Pos{
			Depth: n.Depth + 1,
			FileName: filepath.Join(n.FileName, fileName),
			ParentLine: parentLine,
			CurrentLine: currentLine,
		},
	}
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
			ParentLine: []string{},
		},
	}

	r = WalkDir(v, n, r)

	fmt.Printf("\n%d directories, %d files\n", r.DirNum, r.FileNum)
	return nil
}
