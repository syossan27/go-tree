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
		INodeRecords []INodeRecord
	}

	visitor struct {
	}

	INodeRecord struct {
		INode uint64
		Device int32
	}
)

const (
	ThreeWayLine         = "├── "
	RightAngleLine       = "└── "
	ConnectParentLine    = "│   "
	NonConnectParentLine = "    "
)

func (v visitor) Visit(n Node) Visitor {
	switch NodeType(n.FilePath) {
	case "dir":
		fmt.Print(n.CurrentLine)
		c := color.New(color.FgGreen)
		c.Println(n.FileName)
	case "file":
		fmt.Println(n.CurrentLine + n.FileName)
	case "symlink":
		fmt.Print(n.CurrentLine)
		c := color.New(color.FgMagenta)
		c.Print(n.FileName)
		fmt.Print(" -> ")
		PrintSymlinkRealPath(n.FilePath)
	}
	return v
}

func GetINodeRecord(filePath string) (INodeRecord, bool) {
	var ir INodeRecord
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return ir, false
	}

	// TODO: windows
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return ir, false
	}
	ir.INode = stat.Ino
	ir.Device = stat.Dev

	return ir, true
}

func (r *Result) saveINode(ir INodeRecord) {
	r.INodeRecords = append(r.INodeRecords, ir)
}

func (r Result) searchINode(ir INodeRecord) bool {
	for _, iNodeRecord := range r.INodeRecords {
		if iNodeRecord.INode == ir.INode &&
			iNodeRecord.Device == ir.Device {
			return true
		}
	}
	return false
}

func Walk(c *cli.Context, v Visitor, n Node, r Result) Result {
	// if set L option, stop walk on specify level
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

	switch NodeType(n.FilePath) {
	case "symlink":
		// Plan: symlinkだった時に参照先のパスにnodeのFilePathを上書きして、再度Walk

		// Check symlink linked node type
		realPath, err := os.Readlink(n.FilePath)
		if err != nil {
			break
		}
		switch finalDestinationNodeType(realPath) {
		case "dir":
			r.DirNum++

			// if set l option, recursive symlink
			if !c.Bool("l") {
				fmt.Println("")
			} else {
				iNodeRecord, ok := GetINodeRecord(n.FilePath)
				if ok && r.searchINode(iNodeRecord) {
					fmt.Println("  [recursive, not followed]")
					return r
				} else {
					fmt.Println("")
				}
				r.saveINode(iNodeRecord)

				r = WalkDir(c, v, n, r)
			}
		case "file":
			r.FileNum++
		}
	case "dir":
		r.DirNum++
		r = WalkDir(c, v, n, r)
	case "file":
		r.FileNum++
	}

	return r
}

func WalkDir(c *cli.Context, v Visitor, n Node, r Result) Result {
	files, err := ReadDir(c, n.FilePath)
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

func ReadDir(c *cli.Context, dirPath string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	if !c.Bool("a") {
		files = exceptHiddenFile(files)
	}
	if c.Bool("d") {
		files = exceptFile(dirPath, files)
	}

	return files, nil
}

func exceptHiddenFile(files []os.FileInfo) []os.FileInfo {
	var newFiles []os.FileInfo
	for _, file := range files {
		if !IsHidden(file.Name()) {
			newFiles = append(newFiles, file)
		}
	}
	return newFiles
}

func exceptFile(dirPath string, files []os.FileInfo) []os.FileInfo {
	var newFiles []os.FileInfo
	for _, file := range files {
		if file.IsDir() || IsDirLinkedSymlink(dirPath, file) {
			newFiles = append(newFiles, file)
		}
	}
	return newFiles
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

func NodeType(filePath string) string {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		panic(err)
	}

	if IsSymlinkByFilePath(filePath) {
		return "symlink"
	}

	if fileInfo.IsDir() {
		return "dir"
	} else {
		return "file"
	}
}

func finalDestinationNodeType(realPath string) string {
	var err error
	loop: switch NodeType(realPath) {
	case "dir":
		return "dir"
	case "file":
		return "file"
	case "symlink":
		realPath, err = os.Readlink(realPath)
		if err != nil {
			panic(err)
		}
		goto loop
	}
	return ""
}

func IsSymlinkByFilePath(filePath string) bool {
	fileInfo, err := os.Lstat(filePath)
	if err != nil {
		panic(err)
	}

	if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		return true
	} else {
		return false
	}
}

func IsDirLinkedSymlink(dirPath string, fileInfo os.FileInfo) bool {
	symlinkInfo, err := os.Stat(filepath.Join(dirPath, fileInfo.Name()))
	if err != nil {}
	if symlinkInfo.IsDir() {
		return true
	} else {
		return false
	}
}

func PrintSymlinkRealPath(filePath string) {
	realPath, err := os.Readlink(filePath)
	if err != nil {
		panic(err)
	}

	switch finalDestinationNodeType(realPath) {
	case "dir":
		c := color.New(color.FgGreen)
		c.Print(realPath)
	case "file":
		fmt.Println(realPath)
	}
}

func IsHidden(fileName string) bool {
	if fileName[0:1] == "." {
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

		v := visitor{}
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

			v := visitor{}
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
