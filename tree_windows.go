// +build windows

package go_tree

import (
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

type (
	Visitor interface {
		Visit(n Node) Visitor
	}

	Node struct {
		Pos
	}

	Pos struct {
		Level       int
		FileName 	string
		FilePath    string
		CurrentLine string
		ParentLine  []string
	}

	Result struct {
		DirNum int64
		FileNum int64
	}

	visitor struct {}
)

const (
	ThreeWayLine         = "├── "
	RightAngleLine       = "└── "
	ConnectParentLine    = "│    "
	NonConnectParentLine = "     "
)

func (v visitor) Visit(n Node) Visitor {
	if IsHidden(n.FilePath) {
		return nil
	}

	switch n.Type() {
	case "dir":
		// TODO: coloring print
		fmt.Println(n.CurrentLine + n.FileName)
	case "file":
		// TODO: coloring print
		fmt.Println(n.CurrentLine + n.FileName)
	}
	return v
}

func Walk(c *cli.Context, v Visitor, n Node, r Result) Result {
	levelStr := c.String("L")
	if levelStr != "" {
		level, err := strconv.Atoi(levelStr)
		if err != nil || n.Level > level {
			return r
		}
	}

	if v = v.Visit(n); v == nil {
		return r
	}

	switch n.Type() {
	case "dir":
		r = WalkDir(c, v, n, r)
		r.DirNum++
		return r
	case "file":
		r.FileNum++
		return r
	}

	return r
}

func WalkDir(c *cli.Context, v Visitor, n Node, r Result) Result {
	files, err := ReadDirWithoutHiddenFile(n.FilePath)
	if err != nil {
		return r
	}

	lastIndex := len(files) - 1
	for i, file := range files {
		nextNode := n.NextNode(i, lastIndex, file.Name())
		r = Walk(c, v, nextNode, r)
	}

	return r
}

func ReadDirWithoutHiddenFile(filePath string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		return nil, err
	}

	var withoutHiddenFiles []os.FileInfo
	for _, file := range files {
		if !IsHidden(filepath.Join(filePath, file.Name())) {
			withoutHiddenFiles = append(withoutHiddenFiles, file)
		}
	}

	return withoutHiddenFiles, nil
}

func IsHidden(filePath string) bool {
	p, e := syscall.UTF16PtrFromString(filePath)
	if e != nil {
		return false
	}
	attrs, e := syscall.GetFileAttributes(p)
	if e != nil {
		return false
	}

	return attrs & syscall.FILE_ATTRIBUTE_HIDDEN != 0
}

func (n Node) NextNode(currentIndex, lastIndex int, fileName string) Node {
	var parentLine []string
	currentLine := strings.Join(n.ParentLine, "")
	if currentIndex != lastIndex {
		parentLine = append(n.ParentLine, ConnectParentLine)
		if n.Level != 0 {
			currentLine = currentLine + ThreeWayLine
		} else {
			currentLine = ThreeWayLine
		}
	} else {
		parentLine = append(n.ParentLine, NonConnectParentLine)
		if n.Level != 0 {
			currentLine = currentLine + RightAngleLine
		} else {
			currentLine = RightAngleLine
		}
	}

	return Node{
		Pos{
			Level:       n.Level + 1,
			FileName:    fileName,
			FilePath:    filepath.Join(n.FilePath, fileName),
			ParentLine:  parentLine,
			CurrentLine: currentLine,
		},
	}
}

func (n Node) Type() string {
	fileInfo, err := os.Stat(n.FilePath)
	if err != nil {
		panic(err)
	}

	if fileInfo.IsDir() {
		return "dir"
	} else {
		return "file"
	}
}

func TreeCommand(c *cli.Context) error {
	err := Validate(c)
	if err != nil {
		fmt.Printf("tree: %v.\n", err)
		return nil
	}

	fmt.Println(".")

	rootDir, _ := os.Getwd()

	v := visitor{}
	r := Result{}
	n := Node{
		Pos {
			Level:      0,
			FilePath:   rootDir,
			ParentLine: []string{},
		},
	}

	r = WalkDir(c, v, n, r)

	fmt.Printf("\n%d directories, %d files\n", r.DirNum, r.FileNum)
	return nil
}