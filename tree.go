// +build darwin linux netbsd openbsd freebsd

package go_tree

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
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

	visitor struct {
		HiddenFlag bool
	}
)

const (
	ThreeWayLine         = "├── "
	RightAngleLine       = "└── "
	ConnectParentLine    = "│    "
	NonConnectParentLine = "     "
)

func (v visitor) Visit(n Node) Visitor {
	if !v.HiddenFlag && n.IsHidden() {
		return nil
	}

	switch n.Type() {
	case "dir":
		fmt.Print(n.CurrentLine)
		color.Green(n.FileName)
	case "file":
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
	files, err := ioutil.ReadDir(n.FilePath)
	if err != nil {
		panic(err)
	}

	lastIndex := len(files) - 1
	for i, file := range files {
		nextNode := n.NextNode(i, lastIndex, file.Name())
		r = Walk(c, v, nextNode, r)
	}

	return r
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

func (n Node) IsHidden() bool {
	if n.FileName[0:1] == "." {
		return true
	} else {
		return false
	}
}

func TreeCommand(c *cli.Context) error {
	err := ValidateFlag(c)
	if err != nil {
		fmt.Printf("tree: %v.\n", err)
		return nil
	}

	dirs := c.Args()
	r := Result{}
	if len(dirs) == 0 {
		// Not specify dirs
		fmt.Println(".")

		rootDir, _ := os.Getwd()

		v := visitor{
			HiddenFlag: c.Bool("a"),
		}
		n := Node{
			Pos {
				Level:      0,
				FilePath:   rootDir,
				ParentLine: []string{},
			},
		}

		r = WalkDir(c, v, n, r)
	} else {
		// Specify dirs
		for _, dir := range dirs {
			workingDir, _ := os.Getwd()
			rootDir := filepath.Join(workingDir, dir)

			// validate specify dir
			fileInfo, err := os.Stat(rootDir)
			if  err != nil || !fileInfo.IsDir() {
				fmt.Printf("%s [error opening dir]\n", dir)
				continue
			}

			fmt.Println(dir)

			v := visitor{
				HiddenFlag: c.Bool("a"),
			}
			n := Node{
				Pos {
					Level:      0,
					FilePath:   rootDir,
					ParentLine: []string{},
				},
			}

			r = WalkDir(c, v, n, r)
		}
	}

	fmt.Printf("\n%d directories, %d files\n", r.DirNum, r.FileNum)

	return nil
}
