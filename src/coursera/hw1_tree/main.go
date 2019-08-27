package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var GlobalFiles bool

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, filePath string, printFiles bool) error {
	var treeRoot string
	GlobalFiles = printFiles

	fileList, err := getTreeFileList(filePath)

	for index, file := range fileList {
		currentLine := getLinePath(file)
		if currentLine == "" {
			continue
		}

		treeRoot = treeRoot + currentLine

		if (len(fileList) - 1) != index {
			treeRoot = treeRoot + "\n"
		}
	}

	fmt.Fprintln(out, treeRoot)

	return err
}

//Sorting all files
func getTreeFileList(filePath string) ([]string, error) {
	var fileList []string

	err := filepath.Walk(filePath, func(path string, f os.FileInfo, err error) error {

		if !f.IsDir() && !GlobalFiles {
			return nil
		}

		fileList = append(fileList, path)
		return nil
	})

	return fileList, err
}

//get linePath
func getLinePath(pathOrigin string) string {
	var pathResult string
	var tabs string

	pathLinux := strings.Replace(pathOrigin, `\`, `/`, 100)
	pathListFull := strings.Split(pathLinux, `/`)
	pathList := pathListFull[1:]

	if len(pathList) == 0 {
		return pathResult
	}

	basePath := filepath.Base(pathOrigin) + getFileSize(pathOrigin)

	if isLastElementPath(pathOrigin) {
		pathResult = pathResult + `└───` + basePath
	} else {
		pathResult = pathResult + `├───` + basePath
	}

	tabs = getTabs(pathListFull)

	return tabs + pathResult
}

func getTabs(pathList []string) string {
	var tabResult string

	for i := 2; i < len(pathList); i++ {
		if isLastElementPath(filepath.Join(pathList[:i]...)) {
			tabResult = tabResult + "\t"
		} else {
			tabResult = tabResult + "│\t"
		}
	}

	return tabResult
}

//check for last element
func isLastElementPath(path string) bool {

	basePath := filepath.Base(path)

	var sortList []string

	files, _ := ioutil.ReadDir(filepath.Dir(path))

	for _, file := range files {
		if GlobalFiles == false && file.IsDir() == false {
			continue
		}
		sortList = append(sortList, file.Name())
	}

	if sortList[len(sortList)-1] == basePath {
		return true
	}

	return false
}

//file size
func getFileSize(path string) string {
	var fileSize string
	fileInfo, _ := os.Stat(path)
	if !fileInfo.IsDir() {
		size := fileInfo.Size()
		if size == 0 {
			fileSize = " (empty)"
		} else {
			fileSize = fmt.Sprintf(" (%vb)", size)
		}
	}

	return fileSize
}
