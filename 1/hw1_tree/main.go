package main

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	pathNamesSep = "\\"
	leaf         = "├───"
	endLeaf      = "└───"
)

func print(out io.Writer, path string, stat os.FileInfo, isLastFile bool) {
	var pathForPrint string
	var sizeString string
	for level := strings.Count(path, string(os.PathSeparator)) - 1; level != 0; level-- {
		if !isLastFile {
			pathForPrint += "│"
		}
		pathForPrint += "\t"
	}
	if !isLastFile {
		pathForPrint += leaf
	} else {
		pathForPrint += endLeaf
	}
	pathForPrint += filepath.Base(path)
	if !stat.IsDir() {
		if stat.Size() > 0 {
			sizeString = strconv.Itoa(int(stat.Size())) + "b"
		} else {
			sizeString = "empty"
		}
		pathForPrint += " (" + sizeString + ")"
	}
	pathForPrint += "\n"

	out.Write([]byte(pathForPrint))
}

// PrevDirNames foo bar
type prevDirNames []string

func (prevNames *prevDirNames) push(names []string) {
	if len(names) > 0 {
		*prevNames = append(*prevNames, strings.Join(names, pathNamesSep))
	}
}

func (prevNames *prevDirNames) pop() (names []string) {
	len := len(*prevNames)
	names, *prevNames = strings.Split((*prevNames)[len-1], pathNamesSep), (*prevNames)[:len-1]
	return
}

func prepareNames(names []string, rootPath string) []string {
	newNames := make([]string, len(names))
	for index, path := range names {
		if path == "" {
			continue
		}
		newNames[index] = filepath.Join(rootPath, path)
	}
	return newNames
}

func dirTree(out io.Writer, rootPath string, printFiles bool) (err error) {
	var file *os.File
	var stat os.FileInfo
	var path string
	names := []string{rootPath}
	// Переменная для сохранения иерархии файлов при разборе подпапок
	var prevNames prevDirNames
	isRoot := true
	for len(names) > 0 {
		path, names = names[0], names[1:]
		// Получение файла
		file, err = os.Open(path)
		if err != nil {
			return
		}
		// Получение инфы о файла
		stat, err = file.Stat()
		if err != nil {
			return
		}
		// Eсли директория то получаем новый список вложенных файлов и папок
		// а старый сохраняем
		if stat.IsDir() {
			// Сохранение текущей дериктории
			prevNames.push(names)
			names, err = file.Readdirnames(0)
			if err != nil {
				return
			}
			sort.Strings(names)
			names = prepareNames(names, path)
		}
		// Печать
		if (stat.IsDir() || (!stat.IsDir() && printFiles)) && !isRoot {
			print(out, path, stat, len(names) == 0)
		}
		if isRoot {
			isRoot = false
		}
		file.Close()
		// Если в этой директории кончились файлы то идем на уровень ниже
		if len(names) == 0 && len(prevNames) > 0 {
			names = prevNames.pop()
		}
	}
	return
}

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
