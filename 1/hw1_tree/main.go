package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

const PATH_NAMES_SEP = "\\"

func print(out io.Writer, isDir bool, level int) {

}

func dirTree(out io.Writer, rootPath string, printFiles bool) (err error) {
	var file *os.File
	var stat os.FileInfo
	var path string
	level := 0
	names := []string{rootPath}
	// Переменная для сохранения ф
	var prevNames []string
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
		// Eсли директория то получаем список вложенных файлов и папок
		if stat.IsDir() {
			// Сохранение текущей дериктории
			prevNames = append(prevNames, strings.Join(names, PATH_NAMES_SEP))
			names, err = file.Readdirnames(0)
			if err != nil {
				return
			}
			sort.Strings(names)
		}
		fmt.Printf("%#v\n", names)
		print(out, stat.IsDir(), level)
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
