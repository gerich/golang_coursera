package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
)

const (
	pathNamesSep = "\\"
	leaf         = "├───"
	endLeaf      = "└───"
)

func print(out io.Writer, path dirFiles, currentFiles dirFiles, stat os.FileInfo) {
	var pathForPrint string
	var sizeString string
	higherLevel := path.Len()
	isLastLevel := len(currentFiles) == 0
	for level := higherLevel; level > 0; level-- {
		if !isLastLevel && level > 0 {
			pathForPrint += "│"
		}
		pathForPrint += "\t"
	}

	if !isLastLevel {
		pathForPrint += leaf
	} else {
		pathForPrint += endLeaf
	}
	pathForPrint += stat.Name()

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

func dirPath(rootPath string, paths []os.FileInfo) string {
	path := rootPath
	for _, stat := range paths {
		path += string(os.PathSeparator) + stat.Name()
	}
	return path
}

type dirFiles []os.FileInfo

func (nf dirFiles) Len() int      { return len(nf) }
func (nf dirFiles) Swap(i, j int) { nf[i], nf[j] = nf[j], nf[i] }
func (nf dirFiles) Less(i, j int) bool {
	// Use path names
	pathA := nf[i].Name()
	pathB := nf[j].Name()
	// // Grab integer value of each filename by parsing the string and slicing off
	// // the extension
	// a, err1 := strconv.ParseInt(pathA[0:strings.LastIndex(pathA, ".")], 10, 64)
	// b, err2 := strconv.ParseInt(pathB[0:strings.LastIndex(pathB, ".")], 10, 64)
	// // If any were not numbers sort lexographically
	// if err1 != nil || err2 != nil {
	return pathA < pathB
	// }
	// // Which integer is smaller?
	// return a < b
}

func filterPrinted(stat os.FileInfo, files dirFiles) dirFiles {
	var index int
	var file os.FileInfo
	for index, file = range files {
		if file.Name() == stat.Name() {
			break
		}
	}
	if len(files)-1 == index {
		return files[len(files):]
	}
	return files[index+1:]
}

func dirTree(out io.Writer, rootPath string, printFiles bool) (err error) {
	var prevFiles dirFiles
	var stat os.FileInfo
	var files dirFiles
	files, err = ioutil.ReadDir(rootPath)
	sort.Sort(files)
	if err != nil {
		return
	}
	// Переменная для сохранения иерархии файлов при разборе подпапок
	for len(files) > 0 {
		stat, files = files[0], files[1:]
		// Печать
		if stat.IsDir() || (!stat.IsDir() && printFiles) {
			print(out, prevFiles, files, stat)
		}
		// Eсли директория то получаем новый список вложенных файлов и папок
		// а старый сохраняем
		if stat.IsDir() {
			// Сохранение текущей дериктории
			prevFiles = append(prevFiles, stat)
			files, err = ioutil.ReadDir(dirPath(rootPath, prevFiles))
			if err != nil {
				return
			}
			if !sort.IsSorted(files) {
				sort.Sort(files)
			}
		}
		fmt.Println(stat.Name())
		// Если в этой директории кончились файлы то идем на уровень ниже
		for len(prevFiles) > 0 && len(files) == 0 {
			stat, prevFiles = prevFiles[len(prevFiles)-1], prevFiles[:len(prevFiles)-1]
			files, err = ioutil.ReadDir(dirPath(rootPath, prevFiles))
			if err != nil {
				return
			}
			files = filterPrinted(stat, files)
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
