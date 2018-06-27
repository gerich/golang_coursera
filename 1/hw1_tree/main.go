package main

import (
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

func print(out io.Writer, markers []bool, files dirFiles, stat os.FileInfo, printFiles bool) {
	var pathForPrint string
	var sizeString string
	isLastLevel := len(files) == 0
	if !printFiles {
		isLastLevel = len(filterDirs(files)) == 0
	}
	for level := 0; level < len(markers); level++ {
		if !markers[level] {
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

	if !sort.IsSorted(files) {
		sort.Sort(files)
	}

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

func prepareLastDirMarker(files dirFiles, rootPath string, printFiles bool) (markers []bool, err error) {
	markers = make([]bool, files.Len(), files.Len())
	var filtered, currFiles dirFiles

	for index, stat := range files {
		currFiles, err = ioutil.ReadDir(dirPath(rootPath, files[:index]))
		if !printFiles {
			currFiles = filterDirs(currFiles)
		}
		if err != nil {
			return
		}
		filtered = filterPrinted(stat, currFiles)
		markers[index] = filtered.Len() == 0
	}

	if len(files) > 0 {
		currFiles, err = ioutil.ReadDir(dirPath(rootPath, files[:0]))
		filtered = filterPrinted(files[0], currFiles)
		if !printFiles {
			filtered = filterDirs(filtered)
		}
		markers[0] = filtered.Len() == 0
	}

	return
}

func filterDirs(files dirFiles) dirFiles {
	var filtered dirFiles
	for _, cur := range files {
		if cur.IsDir() {
			filtered = append(filtered, cur)
		}
	}

	return filtered
}

func dirTree(out io.Writer, rootPath string, printFiles bool) (err error) {
	var prevFiles dirFiles
	var stat os.FileInfo
	var files dirFiles
	files, err = ioutil.ReadDir(rootPath)
	sort.Sort(files)
	var markers []bool
	if err != nil {
		return
	}
	// Переменная для сохранения иерархии файлов при разборе подпапок
	for len(files) > 0 {
		stat, files = files[0], files[1:]
		// Печать
		if stat.IsDir() || (!stat.IsDir() && printFiles) {
			markers, err = prepareLastDirMarker(prevFiles, rootPath, printFiles)
			if err != nil {
				return
			}
			print(out, markers, files, stat, printFiles)
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
